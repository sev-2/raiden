package supabase

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hashicorp/go-hclog"
)

var Logger hclog.Logger

func init() {
	Logger = hclog.Default().Named("raiden.pubsub.supabase")
}

const (
	heartbeatInterval = 30 * time.Second
	reconnectBaseWait = 1 * time.Second
	reconnectMaxWait  = 30 * time.Second
	maxReconnectTries = 10
)

// ProviderConfig holds Supabase connection settings (no raiden imports).
type ProviderConfig struct {
	SupabasePublicUrl string
	SupabaseApiUrl    string
	AnonKey           string
	ServiceKey        string
	ProjectId         string
}

// RealtimeChannelType mirrors raiden.RealtimeChannelType without importing root.
type RealtimeChannelType string

const (
	ChannelBroadcast       RealtimeChannelType = "broadcast"
	ChannelPresence        RealtimeChannelType = "presence"
	ChannelPostgresChanges RealtimeChannelType = "postgres_changes"
)

// ListenHandler describes a subscription for the provider.
type ListenHandler struct {
	Name        string
	Topic       string
	ChannelType RealtimeChannelType
	Table       string
	Schema      string
	EventFilter string
	ConsumeFn   func(ctx context.Context, event string, payload json.RawMessage) error
}

// PhoenixMessage is the Phoenix Channels v1 protocol message format.
type PhoenixMessage struct {
	Topic   string          `json:"topic"`
	Event   string          `json:"event"`
	Payload json.RawMessage `json:"payload"`
	Ref     string          `json:"ref,omitempty"`
	JoinRef string          `json:"join_ref,omitempty"`
}

// WebSocketConn abstracts the WebSocket connection for testability.
type WebSocketConn interface {
	WriteJSON(v any) error
	ReadJSON(v any) error
	Close() error
}

// Dialer abstracts WebSocket dialing for testability.
type Dialer interface {
	Dial(url string) (WebSocketConn, error)
}

// Provider implements the Supabase Realtime PubSub provider.
type Provider struct {
	Config *ProviderConfig
	Dialer Dialer

	mu       sync.Mutex
	conn     WebSocketConn
	cancel   context.CancelFunc
	stopped  chan struct{}
	refCount atomic.Int64
}

func (p *Provider) nextRef() string {
	return strconv.FormatInt(p.refCount.Add(1), 10)
}

func (p *Provider) realtimeUrl() (string, error) {
	baseUrl := p.Config.SupabasePublicUrl
	if baseUrl == "" {
		baseUrl = p.Config.SupabaseApiUrl
	}
	if baseUrl == "" {
		return "", fmt.Errorf("supabase realtime: no SupabasePublicUrl or SupabaseApiUrl configured")
	}

	parsed, err := url.Parse(baseUrl)
	if err != nil {
		return "", fmt.Errorf("supabase realtime: invalid base url: %w", err)
	}

	scheme := "wss"
	if parsed.Scheme == "http" {
		scheme = "ws"
	}

	apiKey := p.Config.ServiceKey
	if apiKey == "" {
		apiKey = p.Config.AnonKey
	}

	wsUrl := fmt.Sprintf("%s://%s/realtime/v1/websocket?apikey=%s&vsn=1.0.0",
		scheme, parsed.Host, apiKey)
	return wsUrl, nil
}

func (p *Provider) connect() error {
	wsUrl, err := p.realtimeUrl()
	if err != nil {
		return err
	}

	conn, err := p.Dialer.Dial(wsUrl)
	if err != nil {
		return fmt.Errorf("supabase realtime: dial failed: %w", err)
	}

	p.mu.Lock()
	p.conn = conn
	p.mu.Unlock()

	return nil
}

func (p *Provider) sendJSON(msg any) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.conn == nil {
		return fmt.Errorf("supabase realtime: not connected")
	}
	return p.conn.WriteJSON(msg)
}

func (p *Provider) joinChannel(handler ListenHandler) error {
	topic := p.buildTopic(handler)
	payload := p.buildJoinPayload(handler)

	ref := p.nextRef()
	msg := PhoenixMessage{
		Topic:   topic,
		Event:   "phx_join",
		Payload: payload,
		Ref:     ref,
		JoinRef: ref,
	}

	Logger.Info("joining channel", "topic", topic, "name", handler.Name)
	return p.sendJSON(msg)
}

func (p *Provider) buildTopic(h ListenHandler) string {
	switch h.ChannelType {
	case ChannelPostgresChanges:
		if h.Table != "" {
			return fmt.Sprintf("realtime:%s:%s", h.Schema, h.Table)
		}
		return fmt.Sprintf("realtime:%s", h.Schema)
	default:
		return fmt.Sprintf("realtime:%s", h.Topic)
	}
}

func (p *Provider) buildJoinPayload(h ListenHandler) json.RawMessage {
	config := map[string]any{}

	switch h.ChannelType {
	case ChannelBroadcast:
		config["broadcast"] = map[string]any{"ack": false, "self": false}
	case ChannelPresence:
		config["presence"] = map[string]any{"key": ""}
	case ChannelPostgresChanges:
		pgConfig := map[string]any{
			"event":  h.EventFilter,
			"schema": h.Schema,
		}
		if h.Table != "" {
			pgConfig["table"] = h.Table
		}
		config["postgres_changes"] = []any{pgConfig}
	}

	payload := map[string]any{"config": config}
	b, _ := json.Marshal(payload)
	return b
}

// StartListen connects to Supabase Realtime, joins channels, and dispatches messages.
func (p *Provider) StartListen(handlers []ListenHandler) error {
	if err := p.connect(); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel
	p.stopped = make(chan struct{})

	// Build topic → handler lookup
	topicHandlers := make(map[string]ListenHandler)
	for _, h := range handlers {
		topic := p.buildTopic(h)
		topicHandlers[topic] = h
		if err := p.joinChannel(h); err != nil {
			cancel()
			return fmt.Errorf("supabase realtime: join failed for %s: %w", h.Name, err)
		}
	}

	// Start heartbeat
	go p.heartbeatLoop(ctx)

	// Message read loop
	go func() {
		defer close(p.stopped)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				var msg PhoenixMessage
				p.mu.Lock()
				conn := p.conn
				p.mu.Unlock()

				if conn == nil {
					return
				}

				if err := conn.ReadJSON(&msg); err != nil {
					select {
					case <-ctx.Done():
						return
					default:
						Logger.Error("read error, attempting reconnect", "message", err)
						if reconnErr := p.reconnect(ctx, handlers, topicHandlers); reconnErr != nil {
							Logger.Error("reconnect failed", "message", reconnErr)
							return
						}
						continue
					}
				}

				p.dispatch(ctx, msg, topicHandlers)
			}
		}
	}()

	<-p.stopped
	return nil
}

func (p *Provider) dispatch(ctx context.Context, msg PhoenixMessage, topicHandlers map[string]ListenHandler) {
	// Skip system events
	switch msg.Event {
	case "phx_reply", "phx_close", "heartbeat":
		return
	}

	handler, ok := topicHandlers[msg.Topic]
	if !ok {
		return
	}

	if err := handler.ConsumeFn(ctx, msg.Event, msg.Payload); err != nil {
		Logger.Error("consume error", "name", handler.Name, "event", msg.Event, "message", err)
	}
}

func (p *Provider) heartbeatLoop(ctx context.Context) {
	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			msg := PhoenixMessage{
				Topic:   "phoenix",
				Event:   "heartbeat",
				Payload: json.RawMessage(`{}`),
				Ref:     p.nextRef(),
			}
			if err := p.sendJSON(msg); err != nil {
				Logger.Error("heartbeat failed", "message", err)
			}
		}
	}
}

func (p *Provider) reconnect(ctx context.Context, handlers []ListenHandler, topicHandlers map[string]ListenHandler) error {
	wait := reconnectBaseWait
	for i := 0; i < maxReconnectTries; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}

		Logger.Info("reconnecting", "attempt", i+1)

		// Close old connection
		p.mu.Lock()
		if p.conn != nil {
			_ = p.conn.Close()
			p.conn = nil
		}
		p.mu.Unlock()

		if err := p.connect(); err != nil {
			Logger.Error("reconnect dial failed", "attempt", i+1, "message", err)
			wait = min(wait*2, reconnectMaxWait)
			continue
		}

		// Re-join channels
		for _, h := range handlers {
			if err := p.joinChannel(h); err != nil {
				Logger.Error("rejoin failed", "name", h.Name, "message", err)
			}
		}

		Logger.Info("reconnected successfully")
		return nil
	}
	return fmt.Errorf("supabase realtime: max reconnect attempts (%d) reached", maxReconnectTries)
}

// StopListen closes the WebSocket connection and stops all listeners.
func (p *Provider) StopListen() error {
	if p.cancel != nil {
		p.cancel()
	}

	p.mu.Lock()
	conn := p.conn
	p.conn = nil
	p.mu.Unlock()

	if conn != nil {
		return conn.Close()
	}
	return nil
}

// Publish sends a broadcast message to the specified topic.
func (p *Provider) Publish(ctx context.Context, topic string, message []byte) error {
	payload := map[string]any{
		"event": "message",
		"type":  "broadcast",
	}

	var data any
	if err := json.Unmarshal(message, &data); err != nil {
		payload["payload"] = map[string]any{"data": string(message)}
	} else {
		payload["payload"] = data
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	msg := PhoenixMessage{
		Topic:   fmt.Sprintf("realtime:%s", topic),
		Event:   "broadcast",
		Payload: payloadBytes,
		Ref:     p.nextRef(),
	}

	Logger.Info("publishing broadcast", "topic", topic)
	return p.sendJSON(msg)
}

// CreateSubscription is a no-op for Supabase Realtime (channels are joined on StartListen).
func (p *Provider) CreateSubscription(name, topic string) error {
	return nil
}
