package raiden

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	netclient "github.com/sev-2/raiden/pkg/client/net"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"go.opentelemetry.io/otel/trace"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (r roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return r(req)
}

type stubPubSub struct {
	messages [][]byte
}

type stubJob struct{}

func (stubJob) Name() string                                  { return "stub" }
func (stubJob) Duration() JobDuration                         { return nil }
func (stubJob) After(JobContext, uuid.UUID, string)           {}
func (stubJob) AfterErr(JobContext, uuid.UUID, string, error) {}
func (stubJob) Before(JobContext, uuid.UUID, string)          {}
func (stubJob) Task(JobContext) error                         { return nil }

func (s *stubPubSub) Register(handler SubscriberHandler) {}
func (s *stubPubSub) Publish(_ context.Context, _ PubSubProviderType, _ string, message []byte) error {
	s.messages = append(s.messages, message)
	return nil
}
func (s *stubPubSub) Listen()                                                  {}
func (s *stubPubSub) Serve(SubscriberHandler) (fasthttp.RequestHandler, error) { return nil, nil }
func (s *stubPubSub) Handlers() []SubscriberHandler                            { return nil }

func TestCtxBasicFlows(t *testing.T) {
	original := netclient.GetClient
	t.Cleanup(func() { netclient.GetClient = original })

	netclient.GetClient = func() netclient.Client {
		return &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			body := `{"message":"ok"}`
			resp := &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(body)),
				Header:     make(http.Header),
			}
			return resp, nil
		})}
	}

	cfg := &Config{SupabasePublicUrl: "http://example.com", TraceEnable: true}
	jobChan := make(chan JobParams, 1)
	tracer := trace.NewNoopTracerProvider().Tracer("ctx")

	ctx := NewCtx(cfg, tracer, jobChan)
	ctx.SetCtx(context.Background())
	fasthttpCtx := &fasthttp.RequestCtx{}
	ctx.RequestCtx = fasthttpCtx

	require.Equal(t, cfg, ctx.Config())
	require.Equal(t, tracer, ctx.Tracer())

	_, span := tracer.Start(context.Background(), "test-span")
	t.Cleanup(func() { span.End() })
	ctx.SetSpan(span)
	require.Equal(t, span, ctx.Span())

	ctx.SetJobChan(jobChan)

	ctx.Set("greeting", "hello")
	require.Equal(t, "hello", ctx.Get("greeting"))

	fasthttpCtx.Request.URI().QueryArgs().Set("name", "world")
	require.Equal(t, "world", ctx.GetQuery("name"))

	fasthttpCtx.SetUserValue("id", 42)
	require.Equal(t, 42, ctx.GetParam("id"))

	errResp := ctx.SendError("boom").(*ErrorResponse)
	require.Equal(t, "boom", errResp.Message)

	errResp = ctx.SendErrorWithCode(http.StatusBadRequest, errors.New("invalid")).(*ErrorResponse)
	require.Equal(t, http.StatusBadRequest, errResp.StatusCode)

	require.NoError(t, ctx.SendJson(map[string]string{"status": "ok"}))
	require.Contains(t, string(fasthttpCtx.Response.Body()), "status")

	fasthttpCtx.Response.ResetBody()
	ctx.Write([]byte("custom"))
	require.Equal(t, []byte("custom"), fasthttpCtx.Response.Body())

	ctx.WriteError(errors.New("fatal"))
	require.Contains(t, string(fasthttpCtx.Response.Body()), "fatal")

	ctx.RegisterLibraries(map[string]any{"BaseLibrary": &BaseLibrary{}})
	var lib BaseLibrary
	require.NoError(t, ctx.ResolveLibrary(&lib))

	stub := &stubPubSub{}
	ctx.pubSub = stub
	require.NoError(t, ctx.Publish(context.Background(), PubSubProviderGoogle, "topic", []byte("data")))
	require.Len(t, stub.messages, 1)

	ctx.jobChan = jobChan
	spanCtx := trace.NewSpanContext(trace.SpanContextConfig{TraceID: trace.TraceID{1}, SpanID: trace.SpanID{2}, TraceFlags: trace.FlagsSampled})
	ctx.SetCtx(trace.ContextWithSpanContext(context.Background(), spanCtx))

	jobCtx, err := ctx.NewJobCtx()
	require.NoError(t, err)
	jobCtx.RunJob(JobParams{Job: stubJob{}, Data: JobData{"k": "v"}})

	select {
	case params := <-jobChan:
		require.Equal(t, "v", params.Data["k"])
	case <-time.After(time.Millisecond * 100):
		t.Fatal("job did not run")
	}

	var target struct{ Message string }
	require.NoError(t, ctx.HttpRequestAndBind(http.MethodGet, "http://example.com", nil, nil, time.Second, &target))
	require.Equal(t, "ok", target.Message)

	resp, err := ctx.HttpRequest(http.MethodGet, "http://example.com", nil, nil, time.Second)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	require.Error(t, ctx.HttpRequestAndBind(http.MethodGet, "http://example.com", nil, nil, time.Second, target))
	require.Error(t, (&Ctx{}).Publish(context.Background(), PubSubProviderGoogle, "topic", nil))
}

func TestBaseLibrary(t *testing.T) {
	var base BaseLibrary
	require.False(t, base.IsLongRunning())
}
