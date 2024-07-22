package raiden_test

import (
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/sev-2/raiden"
	"github.com/stretchr/testify/assert"
)

func MockMiddlewareFn() raiden.MiddlewareFn {
	return func(next raiden.RouteHandlerFn) raiden.RouteHandlerFn {
		return nil
	}
}

func TestServer_RegisterRoute(t *testing.T) {
	conf := loadConfig()
	s := raiden.NewServer(conf)
	routes := []*raiden.Route{
		{
			Methods: []string{"GET"}, Path: "/",
		},
	}

	s.RegisterRoute(routes)
	assert.NotNil(t, s.Router)
}

func TestServer_Use(t *testing.T) {
	conf := loadConfig()
	s := raiden.NewServer(conf)
	middleware := MockMiddlewareFn()

	s.Use(middleware)
}

func TestServer_StartStop(t *testing.T) {
	if os.Getenv("TEST_RUN") == "1" {
		conf := loadConfig()
		s := raiden.NewServer(conf)
		s.Run()
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestServer_StartStop")
	cmd.Env = append(os.Environ(), "TEST_RUN=1")
	err := cmd.Start()
	assert.NoError(t, err)

	time.Sleep(1 * time.Second)
	err1 := cmd.Process.Signal(syscall.SIGTERM)
	assert.NoError(t, err1)
}
