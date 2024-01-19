package raiden

import (
	"fmt"
	"log"
	"net/http"
)

// Define App Server
type Server struct {
	Config      *Config
	HttpServer  *http.Server
	Controllers []Controller
}

func NewServer(config *Config) *Server {
	host, port := "127.0.0.1", "8002"
	if config.ServerHost != "" {
		host = config.ServerHost
	}

	if config.ServerPort != "" {
		port = config.ServerPort
	}

	serverAddr := fmt.Sprintf("%s:%s", host, port)

	Info("set addr to : ", serverAddr)
	s := &Server{
		Config: config,
		HttpServer: &http.Server{
			Addr: serverAddr,
		},
	}
	return s
}

func (s *Server) Run() {
	// initial route
	router := NewRouter(s.Config)

	// print available route
	router.PrintRegisteredRoute()

	// setup handler
	s.HttpServer.Handler = router.GetHandler()

	// run server
	Info("running server on ", s.HttpServer.Addr)
	if err := s.HttpServer.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
