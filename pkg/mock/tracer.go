package mock

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/embedded"
)

type ContextKey string

const SpanNameKey ContextKey = "mock-span-name"

type MockTracer struct {
	embedded.Tracer
}

func (m *MockTracer) Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return context.WithValue(ctx, SpanNameKey, spanName), &MockSpan{}
}

type MockSpan struct {
	embedded.Span
	// Fields to track mock behavior and state (optional)

	Name           string
	Attributes     []attribute.KeyValue
	StatusCode     codes.Code
	StatusMessage  string
	RecordingState bool
	SpanCtx        trace.SpanContext
	Events         []string
	Links          []trace.Link
}

func (m *MockSpan) End(options ...trace.SpanEndOption) {
	// Optionally handle span end logic for testing
}

func (m *MockSpan) AddEvent(name string, options ...trace.EventOption) {
	m.Events = append(m.Events, name)
}

func (m *MockSpan) AddLink(link trace.Link) {
	m.Links = append(m.Links, link)
}

func (m *MockSpan) IsRecording() bool {
	return m.RecordingState
}

func (m *MockSpan) RecordError(err error, options ...trace.EventOption) {
	// Optionally record the error for testing
}

func (m *MockSpan) SpanContext() trace.SpanContext {
	return m.SpanCtx
}

func (m *MockSpan) SetStatus(code codes.Code, description string) {
	m.StatusCode = code
	m.StatusMessage = description
}

func (m *MockSpan) SetName(name string) {
	m.Name = name
}

func (m *MockSpan) SetAttributes(kv ...attribute.KeyValue) {
	m.Attributes = append(m.Attributes, kv...)
}

func (m *MockSpan) TracerProvider() trace.TracerProvider {
	return nil // Return nil or mock implementation for testing
}
