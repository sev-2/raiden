package tracer_test

import (
	"context"
	"testing"

	"github.com/sev-2/raiden/pkg/tracer"
	"github.com/stretchr/testify/assert"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// MockSpanExporter is a mock implementation of sdktrace.SpanExporter
type MockSpanExporter struct {
	sdktrace.SpanExporter
	ShutdownFunc func(ctx context.Context) error
}

func (m *MockSpanExporter) Shutdown(ctx context.Context) error {
	if m.ShutdownFunc != nil {
		return m.ShutdownFunc(ctx)
	}
	return nil
}

func TestStartAgent(t *testing.T) {
	agentConfig := tracer.AgentConfig{
		Collector:   tracer.OtplCollector,
		Endpoint:    "http://localhost:4318",
		Environment: "test",
		Name:        "test-service",
		Version:     "1.0.0",
	}

	shutdown, err := tracer.StartAgent(agentConfig)
	assert.NoError(t, err)
	assert.NotNil(t, shutdown)

	err = shutdown(context.Background())
	assert.NoError(t, err)
}

func TestStartAgentNonHttp(t *testing.T) {
	agentConfig := tracer.AgentConfig{
		Collector:   tracer.OtplCollector,
		Endpoint:    "tcp://localhost:4317",
		Environment: "test",
		Name:        "test-service",
		Version:     "1.0.0",
	}

	_, err := tracer.StartAgent(agentConfig)
	assert.Error(t, err)
}
