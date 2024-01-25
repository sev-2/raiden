package raiden

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/tracer"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/reuseport"
	"go.opentelemetry.io/otel"

	go_context "context"
)

// --- server configuration ----
type Server struct {
	Addr        string
	Controllers []Controller
	Context     context
	HttpServer  *fasthttp.Server
	Middlewares []MiddlewareFn

	ShutdownFunc []func(ctx go_context.Context) error
}

func NewServer(config *Config, controllers []Controller) *Server {
	// setup default configuration
	host, port := "127.0.0.1", "8002"
	if config.ServerHost != "" {
		host = config.ServerHost
	}

	if config.ServerPort != "" {
		port = config.ServerPort
	}

	if config.Version != "" {
		config.Version = "1.0.0"
	}

	if config.Environment == "" {
		config.Environment = "development"
	}
	serverAddr := fmt.Sprintf("%s:%s", host, port)

	return &Server{
		Addr: serverAddr,
		Context: context{
			config:   config,
			rContext: go_context.Background(),
		},
		Controllers: controllers,
		HttpServer:  &fasthttp.Server{},
	}
}

func (s *Server) Use(middleware MiddlewareFn) {
	s.Middlewares = append(s.Middlewares, middleware)
}

func (s *Server) Shutdown(ctx go_context.Context) error {
	shutdownCtx, shutdownCancelFn := go_context.WithTimeout(ctx, 10*time.Second)
	defer shutdownCancelFn()

	for _, sf := range s.ShutdownFunc {
		if err := sf(shutdownCtx); err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) ConfigureTracer() {
	Info("configure tracer")
	if !s.Context.Config().TraceEnable {
		return
	}

	tracerConfig := tracer.AgentConfig{
		Name:        s.Context.Config().ProjectName,
		Collector:   tracer.TraceCollector(s.Context.Config().TraceCollector),
		Endpoint:    s.Context.Config().TraceEndpoint,
		Environment: s.Context.Config().Environment,
		Version:     "1.0.0",
	}
	shutdownFn, err := tracer.StartAgent(tracerConfig)
	if err != nil {
		logger.Panic(err)
	}

	Infof(
		"tracer connected to %q with service name %q in environment %q with version %q",
		tracerConfig.Endpoint, tracerConfig.Name, tracerConfig.Environment, tracerConfig.Version,
	)
	s.Middlewares = append(s.Middlewares, TraceMiddleware)
	s.Context.tracer = otel.Tracer(fmt.Sprintf("%s tracer", s.Context.Config().ProjectName))
	s.ShutdownFunc = append(s.ShutdownFunc, shutdownFn)
}

func (s *Server) ConfigureRoute() {
	Info("configure router")
	// initial route
	router := NewRouter(&s.Context)
	router.RegisterMiddlewares(s.Middlewares)
	router.RegisterControllers(s.Controllers)

	// print available route
	router.PrintRegisteredRoute()

	// set handler
	s.HttpServer.Handler = router.GetHandler()
}

func (s *Server) prepareServer() (h string, l net.Listener, errChan chan error) {
	ln, err := reuseport.Listen("tcp4", s.Addr)
	if err != nil {
		Fatalf("error in reuseport listener: %s", err)
	}

	// create a graceful shutdown listener
	duration := 5 * time.Second
	l = NewGracefulListener(ln, duration)

	// Get hostname
	hostname, err := os.Hostname()
	if err != nil {
		Fatalf("hostname unavailable: %s", err)
	}
	h = hostname

	// Error handling
	errChan = make(chan error, 1)
	return
}

func (s *Server) runServer(hostname string, listener net.Listener, errChan chan error) {
	Infof("%s - Server starting on %v", hostname, listener.Addr())
	Infof("%s - Press Ctrl+C to stop", hostname)
	errChan <- s.HttpServer.Serve(listener)
}

func (s *Server) Run() {
	// setup
	s.ConfigureTracer()
	s.ConfigureRoute()

	// prepare server
	h, l, lErrChan := s.prepareServer()

	/// Run server
	go s.runServer(h, l, lErrChan)

	// SIGINT/SIGTERM handling
	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, syscall.SIGINT, syscall.SIGTERM)

	// Handle channels/graceful shutdown
	for {
		select {
		// If server.ListenAndServe() cannot start due to errors such
		// as "port in use" it will return an error.
		case err := <-lErrChan:
			// running server close for clean up all dependency
			Infof("%s - Clean up all dependency resource", h)
			if errShutdown := s.Shutdown(go_context.Background()); errShutdown != nil {
				Warningf("%s - Server shutdown : %s", h, errShutdown)
			}

			if err != nil {
				Fatalf("%s - Listener error: %s", h, err)
			}

			Infof("%s - Server is shutdown bye :)", h)
			os.Exit(0)

		// handle termination signal
		case <-osSignals:
			fmt.Printf("\n")
			Warningf("%s - Shutdown signal received. starting shutdown server ...", h)

			// Servers in the process of shutting down should disable KeepAlives
			// FIXME: This causes a data race
			Infof("%s - Disable keep alive connection", h)
			s.HttpServer.DisableKeepalive = true

			// Attempt the graceful shutdown by closing the listener
			// and completing all inflight requests.
			Infof("%s - Close connection and wait all request done", h)
			if err := l.Close(); err != nil {
				Fatalf("%s - Error with graceful close : %s", h, err)
			}

			Infof("%s - Server gracefully stopped.", h)
		}
	}
}

// --- graceful shutdown listener ----
type GracefulListener struct {
	// inner listener
	ln net.Listener

	// maximum wait time for graceful shutdown
	maxWaitTime time.Duration

	// this channel is closed during graceful shutdown on zero open connections.
	done chan struct{}

	// the number of open connections
	connsCount uint64

	// becomes non-zero when graceful shutdown starts
	shutdown uint64
}

// NewGracefulListener wraps the given listener into 'graceful shutdown' listener.
func NewGracefulListener(ln net.Listener, maxWaitTime time.Duration) net.Listener {
	return &GracefulListener{
		ln:          ln,
		maxWaitTime: maxWaitTime,
		done:        make(chan struct{}),
	}
}

// Accept creates a conn
func (ln *GracefulListener) Accept() (net.Conn, error) {
	c, err := ln.ln.Accept()

	if err != nil {
		return nil, err
	}

	atomic.AddUint64(&ln.connsCount, 1)

	return &gracefulConn{
		Conn: c,
		ln:   ln,
	}, nil
}

// Addr returns the listen address
func (ln *GracefulListener) Addr() net.Addr {
	return ln.ln.Addr()
}

// Close closes the inner listener and waits until all the pending
// open connections are closed before returning.
func (ln *GracefulListener) Close() error {
	err := ln.ln.Close()
	if err != nil {
		return err
	}

	return ln.waitForZeroConns()
}

func (ln *GracefulListener) waitForZeroConns() error {
	atomic.AddUint64(&ln.shutdown, 1)

	if atomic.LoadUint64(&ln.connsCount) == 0 {
		close(ln.done)
		return nil
	}

	select {
	case <-ln.done:
		return nil
	case <-time.After(ln.maxWaitTime):
		return fmt.Errorf("cannot complete graceful shutdown in %s", ln.maxWaitTime)
	}
}

func (ln *GracefulListener) closeConn() {
	connsCount := atomic.AddUint64(&ln.connsCount, ^uint64(0))

	if atomic.LoadUint64(&ln.shutdown) != 0 && connsCount == 0 {
		close(ln.done)
	}
}

type gracefulConn struct {
	net.Conn
	ln *GracefulListener
}

func (c *gracefulConn) Close() error {
	err := c.Conn.Close()

	if err != nil {
		return err
	}

	c.ln.closeConn()

	return nil
}
