package raiden

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/tracer"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/reuseport"
)

var ServerLogger = logger.HcLog().Named("raiden.server")

// --- server configuration ----
type Server struct {
	Config          *Config
	Router          *router
	HttpServer      *fasthttp.Server
	SchedulerServer gocron.Scheduler
	ShutdownFunc    []func(ctx context.Context) error
	jobs            []Job
}

func NewServer(config *Config) *Server {
	return &Server{
		Config:     config,
		Router:     NewRouter(config),
		HttpServer: &fasthttp.Server{},
	}
}

func (s *Server) RegisterRoute(routes []*Route) {
	s.Router.routes = append(s.Router.routes, routes...)
}

func (s *Server) RegisterJobs(jobs []Job) {
	s.jobs = append(s.jobs, jobs...)
}

func (s *Server) Use(middleware MiddlewareFn) {
	s.Router.middlewares = append(s.Router.middlewares, middleware)
}

func (s *Server) Shutdown(ctx context.Context) error {
	shutdownCtx, shutdownCancelFn := context.WithTimeout(ctx, 10*time.Second)
	defer shutdownCancelFn()

	for _, sf := range s.ShutdownFunc {
		if err := sf(shutdownCtx); err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) configureTracer() {
	ServerLogger.Info("configure tracer")
	tracerConfig := tracer.AgentConfig{
		Name:        s.Config.ProjectName,
		Collector:   tracer.TraceCollector(s.Config.TraceCollector),
		Endpoint:    s.Config.TraceCollectorEndpoint,
		Environment: s.Config.Environment,
		Version:     "1.0.0",
	}
	shutdownFn, err := tracer.StartAgent(tracerConfig)
	if err != nil {
		ServerLogger.Error("configure tracer err", "err", err)
	}

	ServerLogger.With("host", tracerConfig.Endpoint).With("name", tracerConfig.Name).With("environment", tracerConfig.Environment).With("version", tracerConfig.Version).
		Info("tracer connected")
	s.ShutdownFunc = append(s.ShutdownFunc, shutdownFn)
}

func (s *Server) configureRoute() {
	ServerLogger.Info("configure router")

	// build router
	s.Router.BuildHandler()

	// print available route
	s.Router.PrintRegisteredRoute()

	// set handler
	s.HttpServer.Handler = s.Router.GetHandler()
}

func (s *Server) configure() {
	if s.Config.TraceEnable {
		s.configureTracer()
	}

	s.configureRoute()
}

func (s *Server) prepareServer() (h string, l net.Listener, errChan chan error) {
	addr := fmt.Sprintf("%s:%s", s.Config.ServerHost, s.Config.ServerPort)
	ln, err := reuseport.Listen("tcp4", addr)
	if err != nil {
		ServerLogger.Error("prepare server", "msf", err)
		os.Exit(1)
	}

	// create a graceful shutdown listener
	duration := 5 * time.Second
	l = NewGracefulListener(ln, duration)

	// Get hostname
	hostname, err := os.Hostname()
	if err != nil {
		ServerLogger.Error("hostname unavailable", "msf", err)
		os.Exit(1)
	}
	h = hostname

	s.HttpServer.DisablePreParseMultipartForm = true

	// Error handling
	errChan = make(chan error, 1)
	return
}

func (s *Server) prepareScheduleServer() {
	if s.Config.ScheduleStatus == ScheduleStatusOff {
		return
	}

	ss, err := NewSchedulerServer(s.Config)
	if err != nil {
		os.Exit(1)
		return
	}
	s.SchedulerServer = ss.Server

	// register job
	if len(s.jobs) > 0 {
		for _, j := range s.jobs {
			if j == nil {
				continue
			}
			if err := ss.RegisterJob(j); err != nil {
				os.Exit(1)
				return
			}
		}
	}

	s.SchedulerServer.Start()
}

func (s *Server) runHttpServer(hostname string, listener net.Listener, errChan chan error) {
	ServerLogger.Info("started server", "hostname", hostname, "addr", listener.Addr())
	ServerLogger.Info("press Ctrl+C to stop")
	errChan <- s.HttpServer.Serve(listener)
}

func (s *Server) Run() {
	s.configure()

	s.prepareScheduleServer()

	// prepare server
	h, l, lErrChan := s.prepareServer()

	/// Run server
	go s.runHttpServer(h, l, lErrChan)

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
			ServerLogger.Info("clean up all dependency resource")
			if errShutdown := s.Shutdown(context.Background()); errShutdown != nil {
				ServerLogger.Warn("server shutdown  error", "msg", errShutdown.Error())
			}

			// shutdown scheduler
			if s.SchedulerServer != nil {
				SchedulerLogger.Info("start shutdown ")
				if err := s.SchedulerServer.Shutdown(); err != nil {
					SchedulerLogger.Error("error shutdown ", "msg", err.Error())
				}
			}

			if err != nil {
				ServerLogger.Error("listener error ", "msg", err.Error())
				os.Exit(1)
			}

			ServerLogger.Info("server is shutdown bye :)")
			os.Exit(0)

		// handle termination signal
		case <-osSignals:
			ServerLogger.Warn("shutdown signal received. starting shutdown server ...")

			// Servers in the process of shutting down should disable KeepAlives
			// FIXME: This causes a data race
			ServerLogger.Info("disable keep alive connection")
			s.HttpServer.DisableKeepalive = true

			// shutdown scheduler=
			if s.SchedulerServer != nil {
				SchedulerLogger.Info("start shutdown ")
				if err := s.SchedulerServer.Shutdown(); err != nil {
					SchedulerLogger.Error("error shutdown ", "msg", err.Error())
				}
			}

			// Attempt the graceful shutdown by closing the listener
			// and completing all inflight requests.
			ServerLogger.Info("close connection and wait all request done")
			if err := l.Close(); err != nil {
				// shutdown scheduler
				ServerLogger.Error("listener error ", "msg", err.Error())
				os.Exit(1)
			}

			ServerLogger.Info("server gracefully stopped.")
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
