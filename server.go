package raiden

import (
	"fmt"
	"log"
	"net/http"

	"github.com/valyala/fasthttp"
)

// Define App Server
type Server struct {
	Addr        string
	Config      *Config
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
	s := &Server{
		Addr:   serverAddr,
		Config: config,
	}

	return s
}

func (s *Server) Run() {
	// initial route
	router := NewRouter(s.Config)

	// print available route
	router.PrintRegisteredRoute()

	// run server
	Info("running server on ", s.Addr)
	if err := fasthttp.ListenAndServe(s.Addr, router.GetHandler()); err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
