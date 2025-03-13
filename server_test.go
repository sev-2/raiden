package raiden_test

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"github.com/hashicorp/go-hclog"
	"github.com/sev-2/raiden"
	"github.com/stretchr/testify/assert"
)

type PrinHello struct {
	raiden.JobBase
}

type LibA struct{}
type LibB struct{}

func (j *PrinHello) Name() string {
	return "print-hello"
}

func (j *PrinHello) After(ctx raiden.JobContext, jobID uuid.UUID, jobName string) {
}

func (j *PrinHello) AfterErr(ctx raiden.JobContext, jobID uuid.UUID, jobName string, err error) {
}

func (j *PrinHello) Before(ctx raiden.JobContext, jobID uuid.UUID, jobName string) {
}

func (j *PrinHello) Duration() gocron.JobDefinition {
	return gocron.DurationJob(time.Second * 20)
}

func (j *PrinHello) Task(ctx raiden.JobContext) error {
	fmt.Printf("[%s] this message executed at %s \n", ctx.Config().ProjectName, time.Now().String())

	return nil
}

type TestSubscriber struct {
	raiden.SubscriberBase
}

func (s *TestSubscriber) Name() string {
	return "Test New"
}

func (s *TestSubscriber) Provider() raiden.PubSubProviderType {
	return raiden.PubSubProviderGoogle
}

func (s *TestSubscriber) Subscripbtion() string {
	return "test.new-sub"
}

func (s *TestSubscriber) Consume(ctx raiden.SubscriberContext, message any) error {
	msg, valid := message.(*pubsub.Message)
	if !valid {
		return errors.New("invalid google pubsub message")
	}

	raiden.Info("test new listener executed", "data", string(msg.Data))
	return nil
}

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

func TestServer_RegisterJob(t *testing.T) {
	conf := loadConfig()
	s := raiden.NewServer(conf)

	s.RegisterJobs(&PrinHello{})
	assert.NotNil(t, s.RegisterJobs)
}

func TestServer_RegisterSubscriber(t *testing.T) {
	conf := loadConfig()
	s := raiden.NewServer(conf)

	s.RegisterSubscribers(&TestSubscriber{})
	assert.NotNil(t, s.RegisterSubscribers)
}

func TestServer_Use(t *testing.T) {
	conf := loadConfig()
	s := raiden.NewServer(conf)
	middleware := MockMiddlewareFn()

	s.Use(middleware)
}

func TestServer_Log(t *testing.T) {
	conf := loadConfig()
	conf.LogLevel = "trace"
	s := raiden.NewServer(conf)
	s.ConfigureLogLevel()
	assert.Equal(t, hclog.Trace, raiden.GetLogLevel())

	conf.LogLevel = "debug"
	s = raiden.NewServer(conf)
	s.ConfigureLogLevel()
	assert.Equal(t, hclog.Debug, raiden.GetLogLevel())

	conf.LogLevel = "warning"
	s = raiden.NewServer(conf)
	s.ConfigureLogLevel()
	assert.Equal(t, hclog.Warn, raiden.GetLogLevel())

	conf.LogLevel = "error"
	s = raiden.NewServer(conf)
	s.ConfigureLogLevel()
	assert.Equal(t, hclog.Error, raiden.GetLogLevel())
}

func TestServer_StartStop(t *testing.T) {
	if os.Getenv("TEST_RUN") == "1" {
		conf := loadConfig()
		s := raiden.NewServer(conf)
		s.RegisterJobs(&PrinHello{})
		s.RegisterSubscribers(&TestSubscriber{})
		s.RegisterLibs(func(config *raiden.Config) any {
			return SomeLib{
				config: config,
			}
		})

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

func TestServer_RegisterLibs(t *testing.T) {
	conf := loadConfig()
	s := raiden.NewServer(conf)

	s.RegisterLibs(func(config *raiden.Config) any {
		return SomeLib{
			config: config,
		}
	})

	assert.NotNil(t, s.RegisterLibs)
}
