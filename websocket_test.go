package raiden_test

import (
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/fasthttp/websocket"
	"github.com/sev-2/raiden"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

type MockConfig struct {
	ServiceKey string
}

func (m *MockConfig) LoadConfig(configFilePath *string) (*MockConfig, error) {
	return &MockConfig{ServiceKey: "test_service_key"}, nil
}

func TestWebSocketHandler(t *testing.T) {
	// Set up a test WebSocket server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Fatalf("Failed to upgrade: %v", err)
		}
		defer conn.Close()

		for {
			mt, message, err := conn.ReadMessage()
			if err != nil {
				return
			}
			err = conn.WriteMessage(mt, message)
			if err != nil {
				return
			}
		}
	}))
	defer server.Close()

	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("Failed to parse URL: %v", err)
	}

	u.Scheme = strings.Replace(u.Scheme, "http", "ws", 1)

	log.Println("Server URL:", u.String())

	// Create a new fasthttp request context
	ctx := &fasthttp.RequestCtx{}
	ctx.Request.Header.Set("Connection", "upgrade")
	ctx.Request.Header.Set("Upgrade", "websocket")
	ctx.Request.Header.Set("Sec-Websocket-Version", "13")
	ctx.Request.Header.Set("Sec-Websocket-Key", "holla")

	// Call the WebSocket handler
	raiden.WebSocketHandler(ctx, u)

	// Verify the WebSocket upgrade response
	assert.Equal(t, fasthttp.StatusSwitchingProtocols, ctx.Response.StatusCode())
}

func TestRealtimeBroadcastHandler(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "test_api_key", r.Header.Get("Apikey"))
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("Failed to parse URL: %v", err)
	}

	ctx := &fasthttp.RequestCtx{}
	ctx.Request.Header.Set("Connection", "upgrade")
	ctx.Request.Header.Set("Apikey", "test_api_key")
	ctx.Request.SetBody([]byte(`{"key":"value"}`))

	raiden.RealtimeBroadcastHandler(ctx, u)
	assert.Equal(t, http.StatusOK, ctx.Response.StatusCode())
}
