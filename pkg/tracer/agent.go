package tracer

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type (
	TraceCollector string

	AgentConfig struct {
		Collector   TraceCollector
		Endpoint    string
		Environment string
		Name        string
		Version     string
	}
)

const (
	OtplCollector TraceCollector = "otpl"
)

var traceProvider *sdktrace.TracerProvider

// start new connection to collector server
// for now only support otpl that using jaeger and prometheus
func StartAgent(c AgentConfig) (func(ctx context.Context) error, error) {
	ctx := context.Background()
	resOptions := resource.WithAttributes(
		semconv.ServiceName(c.Name),
		semconv.ServiceVersion(c.Version),
		attribute.String("environment", c.Environment),
	)

	// create new resource
	res, err := resource.New(ctx, resOptions)
	if err != nil {
		return nil, err
	}

	// create new exporter
	traceExporter, err := createExporter(ctx, c)
	if err != nil {
		return nil, err
	}

	// Create new trace provicer
	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)
	traceProvider = tp
	otel.SetTracerProvider(traceProvider)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return traceProvider.Shutdown, nil
}

// create new exporter base on agent configuration
func createExporter(ctx context.Context, c AgentConfig) (sdktrace.SpanExporter, error) {
	switch c.Collector {
	case OtplCollector:
		return createOtplExported(ctx, c.Endpoint)
	}

	return nil, fmt.Errorf("unsupported collector : %v", string(c.Collector))
}

// create new otpl exporter base on endpoint
// support http and grpc exporter
func createOtplExported(ctx context.Context, endpoint string) (sdktrace.SpanExporter, error) {
	url, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	switch url.Scheme {
	case "http", "https":
		traceExporter, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpoint(endpoint))
		if err != nil {
			return nil, fmt.Errorf("failed to create trace exporter: %w", err)
		}
		return traceExporter, nil
	default:
		ctxTimeout, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		conn, err := grpc.DialContext(ctxTimeout, endpoint,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
		}

		traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
		if err != nil {
			return nil, fmt.Errorf("failed to create trace exporter: %w", err)
		}

		return traceExporter, nil
	}
}
