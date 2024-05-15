package raiden

import (
	"bytes"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/valyala/fasthttp"
)

const (
	maxReconnectAttempts = 5
	pingPeriod           = 30 * time.Second
)

var (
	closeWsOnce sync.Once
)

func WebSocketHandler(ctx *fasthttp.RequestCtx, u *url.URL) {
	proxyWebSocket(ctx, u)
}

func proxyWebSocket(ctx *fasthttp.RequestCtx, u *url.URL) {

	upgrader := websocket.FastHTTPUpgrader{
		ReadBufferSize:   2048,
		WriteBufferSize:  2048,
		HandshakeTimeout: 10,
		CheckOrigin:      func(ctx *fasthttp.RequestCtx) bool { return true },
	}

	// Upgrade HTTP connection to WebSocket
	err := upgrader.Upgrade(ctx, func(connClient *websocket.Conn) {
		defer connClient.Close()

		scheme := "ws"
		if u.Scheme == "https" {
			scheme = "wss"
		}

		targetURL := url.URL{
			Scheme:   scheme,
			Host:     u.Host,
			Path:     "realtime/v1/websocket",
			RawQuery: "apikey=" + getConfig().ServiceKey,
		}

		reconnectAttempts := 0
		for {
			// Connect to the target WebSocket server
			log.Println("Connecting to Server URL", targetURL.String())
			connServer, _, err := websocket.DefaultDialer.Dial(targetURL.String(), nil)
			if err != nil {
				reconnectAttempts++
				if reconnectAttempts > maxReconnectAttempts {
					log.Printf("Exceeded maximum reconnect attempts (%d). Exiting.\n", maxReconnectAttempts)
					return
				}
				log.Printf("Failed to connect to target WebSocket server: %v\n", err)
				time.Sleep(5 * time.Second) // Wait before attempting reconnection
				continue
			}

			// Reset reconnectAttempts counter on successful connection
			reconnectAttempts = 0

			// Handle ping-pong messages
			go pingPongHandler(connClient, connServer)

			// Handle WebSocket communication
			done := make(chan struct{})
			var wg sync.WaitGroup
			wg.Add(2)

			// Proxy messages from client to server
			go proxyMessages(connClient, connServer, done, &wg)

			// Proxy messages from server to client
			go proxyMessages(connServer, connClient, done, &wg)

			wg.Wait()

			// Wait until either side closes the connection
			select {
			case <-done:
				// Client closed the connection
				log.Println("Client closed WebSocket connection, closing server connection...")
				connServer.Close()
				return
			}
		}

	})
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		ctx.Error("WebSocket upgrade error", fasthttp.StatusInternalServerError)
	}
}

func proxyMessages(src, dest *websocket.Conn, done chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		mt, msg, err := src.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v\n", err)
			break
		}
		log.Printf("Received message: %s\n", string(msg))

		err = dest.WriteMessage(mt, msg)
		if err != nil {
			log.Printf("Error writing message: %v\n", err)
			break
		}
		log.Printf("Send message: %s", string(msg))
	}

	// Close the 'done' channel exactly once to signal completion
	closeOnce(done)
}

func pingPongHandler(src, dest *websocket.Conn) {
	ticker := time.NewTicker(pingPeriod)

	defer ticker.Stop()

	heartbeat := `{"topic":"phoenix","event":"heartbeat","payload":{},"ref":""}}`

	for {
		select {
		case <-ticker.C:
			err := src.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(time.Second))
			if err != nil {
				log.Printf("Error sending ping to client: %v\n", err)
				return
			}

			err = dest.WriteControl(websocket.PingMessage, []byte(heartbeat), time.Now().Add(time.Second))
			if err != nil {
				log.Printf("Error sending ping to server: %v\n", err)
				return
			}
		}
	}
}

func getConfig() *Config {
	currentDir, err := os.Getwd()
	if err != nil {
		log.Println(err)
		return nil
	}

	configFilePath := strings.Join([]string{currentDir, "app.yaml"}, string(os.PathSeparator))

	config, err := LoadConfig(&configFilePath)
	if err != nil {
		log.Println(err)
		return nil
	}

	return config
}

func closeOnce(ch chan struct{}) {
	closeWsOnce.Do(func() {
		close(ch)
	})
}

func RealtimeBroadcastHandler(ctx *fasthttp.RequestCtx, u *url.URL) {

	supabaseUrl := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   "realtime/v1/api/broadcast",
	}

	client := &http.Client{}

	req, err := http.NewRequest("POST", supabaseUrl.String(), bytes.NewBuffer(ctx.PostBody()))
	if err != nil {
		log.Println(err)
		return
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Apikey", string(ctx.Request.Header.Peek("Apikey")))
	if len(ctx.Request.Header.Peek("Authorization")) > 0 {
		req.Header.Add("Authorization", string(ctx.Request.Header.Peek("Authorization")))
	}

	_, err = client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
}
