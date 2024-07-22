package raiden

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"github.com/sev-2/raiden/pkg/logger"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var SchedulerLogger = logger.HcLog().Named("raiden.scheduler")

// ----- Custom Type
type ScheduleStatus string
type JobData map[string]any
type JobParams struct {
	Job     Job
	Data    JobData
	TraceId string
	SpanId  string
}

const (
	ScheduleStatusOn  ScheduleStatus = "on"
	ScheduleStatusOff ScheduleStatus = "off"
)

// ---- Scheduler context
type JobContext interface {
	context.Context
	SetContext(c context.Context)
	Config() *Config
	RunJob(JobParams)
	Get(key string) any
	Set(key string, value any)
	IsDataExist(key string) bool
	Span() trace.Span
	SetSpan(span trace.Span)
}

func newJobCtx(cfg *Config, jobChan chan JobParams, data JobData) JobContext {
	if data == nil {
		data = make(JobData, 0)
	}
	return &jobContext{
		Context: context.Background(),
		cfg:     cfg,
		jobChan: jobChan,
		data:    data,
	}
}

type jobContext struct {
	context.Context
	cfg     *Config
	jobChan chan JobParams
	data    JobData
	span    trace.Span
}

func (ctx *jobContext) SetContext(c context.Context) {
	ctx.Context = c
}

func (ctx *jobContext) Config() *Config {
	return ctx.cfg
}

func (ctx *jobContext) RunJob(params JobParams) {
	if ctx.cfg.TraceEnable {
		spanCtx := trace.SpanContextFromContext(ctx.Context)
		if spanCtx.HasTraceID() {
			params.TraceId = spanCtx.TraceID().String()
		}

		if spanCtx.HasSpanID() {
			params.SpanId = spanCtx.SpanID().String()
		}
	}
	ctx.jobChan <- params
}

func (ctx *jobContext) Get(key string) any {
	value := ctx.data[key]
	return value
}

func (ctx *jobContext) IsDataExist(key string) bool {
	value := ctx.data[key]
	return value != nil
}

func (ctx *jobContext) Set(key string, value any) {
	ctx.data[key] = value
}

func (ctx *jobContext) Span() trace.Span {
	return ctx.span
}

func (ctx *jobContext) SetSpan(span trace.Span) {
	ctx.span = span
}

// ----- Scheduler Base
type JobDuration = gocron.JobDefinition
type Job interface {
	Name() string
	Duration() JobDuration
	After(ctx JobContext, jobID uuid.UUID, jobName string)
	AfterErr(ctx JobContext, jobID uuid.UUID, jobName string, err error)
	Before(ctx JobContext, jobID uuid.UUID, jobName string)
	Task(ctx JobContext) error
}

type JobBase struct{}

func (j *JobBase) Duration() JobDuration {
	return nil
}

func (j *JobBase) After(ctx JobContext, jobID uuid.UUID, jobName string) {}

func (j *JobBase) AfterErr(ctx JobContext, jobID uuid.UUID, jobName string, err error) {}

func (j *JobBase) Before(ctx JobContext, jobID uuid.UUID, jobName string) {}

func (j *JobBase) Task(ctx JobContext) error {
	return nil
}

// ---- Scheduler monitor
type schedulerMonitor struct {
}

func (m *schedulerMonitor) IncrementJob(id uuid.UUID, name string, tags []string, status gocron.JobStatus) {
	if !strings.HasPrefix(name, "wrapper-executor") {
		logger.HcLog().Info("record job status", "job_name", name, "status", status)
	}
}

func (m *schedulerMonitor) RecordJobTiming(startTime, endTime time.Time, id uuid.UUID, name string, tags []string) {
	if !strings.HasPrefix(name, "wrapper-executor") {
		logger.HcLog().Info("record job time", "job_name", name, "star_time", startTime.Format(time.RFC3339), "end_time", endTime.Format(time.RFC3339), "duration", endTime.Sub(startTime).String())
	}
}

// ----- Scheduler server
type Scheduler interface {
	RegisterJob(job Job) error
	Start()
	Stop(ctx context.Context) error
	ListenJobChan()
	SetTracer(tracer trace.Tracer)
}

func NewSchedulerServer(config *Config, options ...gocron.SchedulerOption) (*SchedulerServer, error) {
	server, err := gocron.NewScheduler(options...)
	if err != nil {
		SchedulerLogger.Error(err.Error())
		return nil, err
	}

	jobChan := make(chan JobParams)

	return &SchedulerServer{
		Config:  config,
		Server:  server,
		JobChan: jobChan,
	}, nil

}

type SchedulerServer struct {
	Config  *Config
	Server  gocron.Scheduler
	jobs    []Job
	tracer  trace.Tracer
	JobChan chan JobParams
}

func (s *SchedulerServer) SetTracer(tracer trace.Tracer) {
	s.tracer = tracer
}

func (s *SchedulerServer) RegisterJob(job Job) error {
	if job == nil {
		return fmt.Errorf("Could not register nil job")
	}

	s.jobs = append(s.jobs, job)

	if job.Duration() != nil {
		j, err := s.Server.NewJob(job.Duration(),
			gocron.NewTask(func(scServer *SchedulerServer, jType reflect.Type) {
				jValue := reflect.New(jType).Interface()
				if jobValue, ok := jValue.(Job); ok {
					jobCtx := newJobCtx(scServer.Config, scServer.JobChan, make(JobData))
					// start tracer
					if scServer.tracer != nil {
						spanCtx, span := scServer.tracer.Start(context.Background(), fmt.Sprintf("job - %s", jobValue.Name()))
						jobCtx.SetContext(spanCtx)
						jobCtx.SetSpan(span)
					}

					_, err := scServer.Server.NewJob(gocron.OneTimeJob(gocron.OneTimeJobStartImmediately()), wrapJobTask(jobCtx, jobValue), getJobOptions(s.Server, jobCtx, jobValue)...)
					if err != nil {
						SchedulerLogger.Error("failed run job", "name", job.Name())
					}
				}
			}, s, reflect.TypeOf(job).Elem()),
			gocron.WithName(fmt.Sprintf("wrapper-executor %s", job.Name())))
		if err != nil {
			SchedulerLogger.Error("failed run job", "name", job.Name())
			return err
		}
		SchedulerLogger.Info("register job", "id", j.ID(), "name", job.Name())
	}

	return nil
}

func (s *SchedulerServer) ListenJobChan() {
	for jobParams := range s.JobChan {
		var job Job
		for _, j := range s.jobs {
			if j.Name() == jobParams.Job.Name() {
				job = j
				break
			}
		}

		if job != nil {
			go func(server gocron.Scheduler, cfg *Config, jobChan chan JobParams, params JobParams) {
				jobCtx := newJobCtx(s.Config, s.JobChan, make(JobData))
				// start tracer
				if s.tracer != nil {
					spanCtx, span := extractTraceJobParam(context.Background(), s.tracer, params)
					jobCtx.SetContext(spanCtx)
					jobCtx.SetSpan(span)
				}
				_, err := server.NewJob(gocron.OneTimeJob(gocron.OneTimeJobStartImmediately()), wrapJobTask(jobCtx, job), getJobOptions(s.Server, jobCtx, params.Job)...)
				if err != nil {
					SchedulerLogger.Error("failed schedule job", "name", job.Name())
				}
			}(s.Server, s.Config, s.JobChan, jobParams)
		}
	}
}

func (s SchedulerServer) Start() {
	s.Server.Start()
}

func (s SchedulerServer) Stop(ctx context.Context) error {
	err := s.Server.Shutdown()
	if err != nil {
		return err
	}

	close(s.JobChan)

	return nil
}

func wrapJobTask(jobCtx JobContext, job Job) gocron.Task {
	return gocron.NewTask(func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("%v", r)
			}
		}()
		err = job.Task(jobCtx)
		return
	})
}

func getJobOptions(server gocron.Scheduler, jobCtx JobContext, job Job) []gocron.JobOption {
	options := make([]gocron.JobOption, 0)

	// setup job name
	options = append(options, gocron.WithName(job.Name()))

	// setup job event listener
	options = append(options, gocron.WithEventListeners(
		gocron.BeforeJobRuns(func(jobID uuid.UUID, jobName string) {
			job.Before(jobCtx, jobID, jobName)
		}),
		gocron.AfterJobRuns(func(jobID uuid.UUID, jobName string) {
			job.After(jobCtx, jobID, jobName)
			err := server.RemoveJob(jobID)
			if err != nil {
				SchedulerLogger.Error("failed removing job", "name", jobName)
			}

			if span := jobCtx.Span(); span != nil {
				span.SetStatus(codes.Ok, fmt.Sprintf("job %s is successfully run", jobName))
				span.End()
			}
		}),
		gocron.AfterJobRunsWithError(func(jobID uuid.UUID, jobName string, err error) {
			job.AfterErr(jobCtx, jobID, jobName, err)
			err1 := server.RemoveJob(jobID)
			if err1 != nil {
				SchedulerLogger.Error("failed removing job", "name", jobName)
			}

			if span := jobCtx.Span(); span != nil {
				span.SetStatus(codes.Error, fmt.Sprintf("job %s is fail", jobName))
				span.RecordError(err)
				span.End()
			}
		}),
	))

	return options
}

func extractTraceJobParam(ctx context.Context, tracer trace.Tracer, params JobParams) (rCtx context.Context, span trace.Span) {
	spanName := fmt.Sprintf("job - %s", params.Job.Name())
	if params.TraceId == "" {
		return tracer.Start(ctx, spanName)
	}

	var spanContextConfig trace.SpanContextConfig
	spanContextConfig.TraceID, _ = trace.TraceIDFromHex(params.TraceId)
	spanContextConfig.SpanID, _ = trace.SpanIDFromHex(params.SpanId)
	spanContextConfig.TraceFlags = 01
	spanContextConfig.Remote = true

	spanContext := trace.NewSpanContext(spanContextConfig)
	traceCtx := trace.ContextWithSpanContext(ctx, spanContext)
	return tracer.Start(traceCtx, spanName)
}
