package supabase

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"
)

// --- Mock WebSocket ---

type mockConn struct {
	mu       sync.Mutex
	written  []any
	readMsgs []PhoenixMessage
	readIdx  int
	closed   bool
	readErr  error
}

func (c *mockConn) WriteJSON(v any) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return fmt.Errorf("connection closed")
	}
	c.written = append(c.written, v)
	return nil
}

func (c *mockConn) ReadJSON(v any) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.readErr != nil {
		return c.readErr
	}
	if c.closed {
		return fmt.Errorf("connection closed")
	}
	if c.readIdx >= len(c.readMsgs) {
		// Block until closed
		c.mu.Unlock()
		time.Sleep(50 * time.Millisecond)
		c.mu.Lock()
		if c.closed {
			return fmt.Errorf("connection closed")
		}
		return fmt.Errorf("no more messages")
	}

	msg := c.readMsgs[c.readIdx]
	c.readIdx++

	b, _ := json.Marshal(msg)
	return json.Unmarshal(b, v)
}

func (c *mockConn) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closed = true
	return nil
}

type mockDialer struct {
	conn *mockConn
	err  error
	url  string
}

func (d *mockDialer) Dial(url string) (WebSocketConn, error) {
	d.url = url
	if d.err != nil {
		return nil, d.err
	}
	return d.conn, nil
}

// --- Tests ---

func TestRealtimeUrl(t *testing.T) {
	tests := []struct {
		name      string
		config    ProviderConfig
		wantUrl   string
		wantError bool
	}{
		{
			name: "uses public url with service key",
			config: ProviderConfig{
				SupabasePublicUrl: "https://abc.supabase.co",
				ServiceKey:        "svc-key",
			},
			wantUrl: "wss://abc.supabase.co/realtime/v1/websocket?apikey=svc-key&vsn=1.0.0",
		},
		{
			name: "falls back to api url",
			config: ProviderConfig{
				SupabaseApiUrl: "https://api.example.com",
				AnonKey:        "anon-key",
			},
			wantUrl: "wss://api.example.com/realtime/v1/websocket?apikey=anon-key&vsn=1.0.0",
		},
		{
			name: "http scheme uses ws",
			config: ProviderConfig{
				SupabasePublicUrl: "http://localhost:54321",
				AnonKey:           "local-key",
			},
			wantUrl: "ws://localhost:54321/realtime/v1/websocket?apikey=local-key&vsn=1.0.0",
		},
		{
			name:      "no url configured",
			config:    ProviderConfig{},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provider{Config: &tt.config}
			got, err := p.realtimeUrl()
			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.wantUrl {
				t.Errorf("got %s, want %s", got, tt.wantUrl)
			}
		})
	}
}

func TestBuildTopic(t *testing.T) {
	p := &Provider{}

	tests := []struct {
		name    string
		handler ListenHandler
		want    string
	}{
		{
			name:    "broadcast",
			handler: ListenHandler{Topic: "chat-room", ChannelType: ChannelBroadcast},
			want:    "realtime:chat-room",
		},
		{
			name:    "presence",
			handler: ListenHandler{Topic: "lobby", ChannelType: ChannelPresence},
			want:    "realtime:lobby",
		},
		{
			name: "postgres changes with table",
			handler: ListenHandler{
				ChannelType: ChannelPostgresChanges,
				Schema:      "public",
				Table:       "messages",
			},
			want: "realtime:public:messages",
		},
		{
			name: "postgres changes schema only",
			handler: ListenHandler{
				ChannelType: ChannelPostgresChanges,
				Schema:      "public",
			},
			want: "realtime:public",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.buildTopic(tt.handler)
			if got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestBuildJoinPayload(t *testing.T) {
	p := &Provider{}

	tests := []struct {
		name    string
		handler ListenHandler
		check   func(t *testing.T, payload map[string]any)
	}{
		{
			name:    "broadcast config",
			handler: ListenHandler{ChannelType: ChannelBroadcast},
			check: func(t *testing.T, payload map[string]any) {
				config, ok := payload["config"].(map[string]any)
				if !ok {
					t.Fatal("missing config")
				}
				bc, ok := config["broadcast"].(map[string]any)
				if !ok {
					t.Fatal("missing broadcast config")
				}
				if bc["ack"] != false || bc["self"] != false {
					t.Errorf("unexpected broadcast config: %v", bc)
				}
			},
		},
		{
			name:    "presence config",
			handler: ListenHandler{ChannelType: ChannelPresence},
			check: func(t *testing.T, payload map[string]any) {
				config, ok := payload["config"].(map[string]any)
				if !ok {
					t.Fatal("missing config")
				}
				_, ok = config["presence"].(map[string]any)
				if !ok {
					t.Fatal("missing presence config")
				}
			},
		},
		{
			name: "postgres changes config",
			handler: ListenHandler{
				ChannelType: ChannelPostgresChanges,
				Schema:      "public",
				Table:       "users",
				EventFilter: "INSERT",
			},
			check: func(t *testing.T, payload map[string]any) {
				config, ok := payload["config"].(map[string]any)
				if !ok {
					t.Fatal("missing config")
				}
				changes, ok := config["postgres_changes"].([]any)
				if !ok || len(changes) == 0 {
					t.Fatal("missing postgres_changes config")
				}
				pc := changes[0].(map[string]any)
				if pc["schema"] != "public" || pc["table"] != "users" || pc["event"] != "INSERT" {
					t.Errorf("unexpected postgres_changes config: %v", pc)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw := p.buildJoinPayload(tt.handler)
			var payload map[string]any
			if err := json.Unmarshal(raw, &payload); err != nil {
				t.Fatalf("failed to unmarshal payload: %v", err)
			}
			tt.check(t, payload)
		})
	}
}

func TestConnect(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		conn := &mockConn{}
		dialer := &mockDialer{conn: conn}
		p := &Provider{
			Config: &ProviderConfig{
				SupabasePublicUrl: "https://test.supabase.co",
				ServiceKey:        "key",
			},
			Dialer: dialer,
		}

		err := p.connect()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.conn != conn {
			t.Error("connection not set")
		}
		if dialer.url != "wss://test.supabase.co/realtime/v1/websocket?apikey=key&vsn=1.0.0" {
			t.Errorf("unexpected dial url: %s", dialer.url)
		}
	})

	t.Run("dial error", func(t *testing.T) {
		dialer := &mockDialer{err: fmt.Errorf("connection refused")}
		p := &Provider{
			Config: &ProviderConfig{
				SupabasePublicUrl: "https://test.supabase.co",
				ServiceKey:        "key",
			},
			Dialer: dialer,
		}

		err := p.connect()
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestPublish(t *testing.T) {
	conn := &mockConn{}
	p := &Provider{
		Config: &ProviderConfig{
			SupabasePublicUrl: "https://test.supabase.co",
			ServiceKey:        "key",
		},
		Dialer: &mockDialer{conn: conn},
	}
	p.conn = conn

	err := p.Publish(context.Background(), "chat-room", []byte(`{"text":"hello"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(conn.written) != 1 {
		t.Fatalf("expected 1 message written, got %d", len(conn.written))
	}

	msg, ok := conn.written[0].(PhoenixMessage)
	if !ok {
		t.Fatal("written message is not PhoenixMessage")
	}
	if msg.Topic != "realtime:chat-room" {
		t.Errorf("unexpected topic: %s", msg.Topic)
	}
	if msg.Event != "broadcast" {
		t.Errorf("unexpected event: %s", msg.Event)
	}
}

func TestStopListen(t *testing.T) {
	conn := &mockConn{}
	p := &Provider{
		Config: &ProviderConfig{},
		conn:   conn,
		cancel: func() {},
	}

	err := p.StopListen()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	conn.mu.Lock()
	closed := conn.closed
	conn.mu.Unlock()

	if !closed {
		t.Error("connection was not closed")
	}
}

func TestCreateSubscription(t *testing.T) {
	p := &Provider{}
	err := p.CreateSubscription("test", "topic")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDispatch(t *testing.T) {
	var receivedEvent string
	var receivedPayload json.RawMessage

	handler := ListenHandler{
		Name:        "test-handler",
		Topic:       "room",
		ChannelType: ChannelBroadcast,
		ConsumeFn: func(ctx context.Context, event string, payload json.RawMessage) error {
			receivedEvent = event
			receivedPayload = payload
			return nil
		},
	}

	p := &Provider{}
	topicHandlers := map[string]ListenHandler{
		"realtime:room": handler,
	}

	payload, _ := json.Marshal(map[string]string{"data": "hello"})
	msg := PhoenixMessage{
		Topic:   "realtime:room",
		Event:   "broadcast",
		Payload: payload,
	}

	p.dispatch(context.Background(), msg, topicHandlers)

	if receivedEvent != "broadcast" {
		t.Errorf("expected event 'broadcast', got '%s'", receivedEvent)
	}
	if string(receivedPayload) != string(payload) {
		t.Errorf("payload mismatch: got %s", string(receivedPayload))
	}
}

func TestDispatchSkipsSystemEvents(t *testing.T) {
	called := false
	handler := ListenHandler{
		ConsumeFn: func(ctx context.Context, event string, payload json.RawMessage) error {
			called = true
			return nil
		},
	}

	p := &Provider{}
	topicHandlers := map[string]ListenHandler{"realtime:room": handler}

	for _, event := range []string{"phx_reply", "phx_close", "heartbeat"} {
		p.dispatch(context.Background(), PhoenixMessage{
			Topic: "realtime:room",
			Event: event,
		}, topicHandlers)
	}

	if called {
		t.Error("system events should be skipped")
	}
}

func TestHeartbeat(t *testing.T) {
	conn := &mockConn{}
	p := &Provider{conn: conn}

	ctx, cancel := context.WithCancel(context.Background())

	// Run heartbeat briefly
	go p.heartbeatLoop(ctx)
	time.Sleep(100 * time.Millisecond)
	cancel()

	// Heartbeat may or may not have fired in 100ms (interval is 30s)
	// Just verify it doesn't panic and cancel works
}

func TestJoinChannel(t *testing.T) {
	conn := &mockConn{}
	p := &Provider{conn: conn}

	tests := []struct {
		name    string
		handler ListenHandler
	}{
		{
			name:    "broadcast",
			handler: ListenHandler{Name: "bc", Topic: "chat", ChannelType: ChannelBroadcast},
		},
		{
			name:    "presence",
			handler: ListenHandler{Name: "pr", Topic: "lobby", ChannelType: ChannelPresence},
		},
		{
			name: "postgres changes",
			handler: ListenHandler{
				Name: "pg", ChannelType: ChannelPostgresChanges,
				Schema: "public", Table: "orders", EventFilter: "INSERT",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := p.joinChannel(tt.handler)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}

	if len(conn.written) != 3 {
		t.Fatalf("expected 3 messages written, got %d", len(conn.written))
	}

	// Verify all are phx_join
	for _, w := range conn.written {
		msg := w.(PhoenixMessage)
		if msg.Event != "phx_join" {
			t.Errorf("expected phx_join event, got %s", msg.Event)
		}
	}
}

func TestJoinChannelNoConnection(t *testing.T) {
	p := &Provider{} // no conn set
	err := p.joinChannel(ListenHandler{Name: "test", Topic: "t", ChannelType: ChannelBroadcast})
	if err == nil {
		t.Error("expected error when not connected")
	}
}

func TestStartListenAndDispatch(t *testing.T) {
	payload, _ := json.Marshal(map[string]string{"text": "hello"})
	broadcastMsg := PhoenixMessage{
		Topic:   "realtime:chat",
		Event:   "broadcast",
		Payload: payload,
	}

	conn := &mockConn{
		readMsgs: []PhoenixMessage{broadcastMsg},
	}

	consumeCalled := make(chan bool, 1)
	handler := ListenHandler{
		Name:        "bc-handler",
		Topic:       "chat",
		ChannelType: ChannelBroadcast,
		ConsumeFn: func(ctx context.Context, event string, p json.RawMessage) error {
			consumeCalled <- true
			return nil
		},
	}

	dialer := &mockDialer{conn: conn}
	p := &Provider{
		Config: &ProviderConfig{
			SupabasePublicUrl: "https://test.supabase.co",
			ServiceKey:        "key",
		},
		Dialer: dialer,
	}

	// StartListen blocks, run in goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- p.StartListen([]ListenHandler{handler})
	}()

	// Wait for consume or timeout
	select {
	case <-consumeCalled:
		// success — handler was called
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for handler to be called")
	}

	// Stop
	_ = p.StopListen()
}

func TestDispatchConsumeError(t *testing.T) {
	handler := ListenHandler{
		Name:        "err-handler",
		Topic:       "room",
		ChannelType: ChannelBroadcast,
		ConsumeFn: func(ctx context.Context, event string, payload json.RawMessage) error {
			return fmt.Errorf("consume failed")
		},
	}

	p := &Provider{}
	topicHandlers := map[string]ListenHandler{"realtime:room": handler}

	payload, _ := json.Marshal(map[string]string{"data": "test"})
	msg := PhoenixMessage{Topic: "realtime:room", Event: "broadcast", Payload: payload}

	// Should not panic, just log the error
	p.dispatch(context.Background(), msg, topicHandlers)
}

func TestDispatchUnknownTopic(t *testing.T) {
	p := &Provider{}
	topicHandlers := map[string]ListenHandler{}

	msg := PhoenixMessage{Topic: "realtime:unknown", Event: "broadcast"}
	// Should not panic — unknown topic is silently ignored
	p.dispatch(context.Background(), msg, topicHandlers)
}

func TestReconnect(t *testing.T) {
	reconnectConn := &mockConn{}
	dialer := &mockDialer{conn: reconnectConn}

	p := &Provider{
		Config: &ProviderConfig{
			SupabasePublicUrl: "https://test.supabase.co",
			ServiceKey:        "key",
		},
		Dialer: dialer,
	}

	handler := ListenHandler{
		Name:        "reconn-handler",
		Topic:       "room",
		ChannelType: ChannelBroadcast,
	}
	topicHandlers := map[string]ListenHandler{"realtime:room": handler}

	ctx := context.Background()
	err := p.reconnect(ctx, []ListenHandler{handler}, topicHandlers)
	if err != nil {
		t.Fatalf("reconnect failed: %v", err)
	}

	if p.conn != reconnectConn {
		t.Error("connection not updated after reconnect")
	}
}

func TestReconnectDialFailure(t *testing.T) {
	dialer := &mockDialer{err: fmt.Errorf("connection refused")}
	p := &Provider{
		Config: &ProviderConfig{
			SupabasePublicUrl: "https://test.supabase.co",
			ServiceKey:        "key",
		},
		Dialer: dialer,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := p.reconnect(ctx, nil, nil)
	if err == nil {
		t.Error("expected error on failed reconnect")
	}
}

func TestStopListenNoConnection(t *testing.T) {
	p := &Provider{cancel: func() {}}
	err := p.StopListen()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStopListenNoCancel(t *testing.T) {
	p := &Provider{}
	err := p.StopListen()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPublishNotConnected(t *testing.T) {
	p := &Provider{Config: &ProviderConfig{}}
	err := p.Publish(context.Background(), "topic", []byte("data"))
	if err == nil {
		t.Error("expected error when not connected")
	}
}

func TestPublishNonJSON(t *testing.T) {
	conn := &mockConn{}
	p := &Provider{Config: &ProviderConfig{}, conn: conn}

	err := p.Publish(context.Background(), "topic", []byte("not json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(conn.written) != 1 {
		t.Fatalf("expected 1 message, got %d", len(conn.written))
	}
}

func TestSendJSONNotConnected(t *testing.T) {
	p := &Provider{}
	err := p.sendJSON(map[string]string{"test": "value"})
	if err == nil {
		t.Error("expected error when not connected")
	}
}
