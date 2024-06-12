package raiden

import (
	"fmt"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"github.com/sev-2/raiden/pkg/logger"
)

var SchedulerLogger = logger.HcLog().Named("raiden.scheduler")

// ----- Custom Type
type ScheduleStatus string

const (
	ScheduleStatusOn  ScheduleStatus = "on"
	ScheduleStatusOff ScheduleStatus = "off"
)

// ----- Scheduler Base
type JobDuration = gocron.JobDefinition
type Job interface {
	Name() string
	Duration() JobDuration
	After(cfg *Config, jobID uuid.UUID, jobName string)
	AfterErr(cfg *Config, jobID uuid.UUID, jobName string, err error)
	Before(cfg *Config, jobID uuid.UUID, jobName string)
	Task(cfg *Config) error
}

type JobBase struct{}

func (j *JobBase) Duration() JobDuration {
	return nil
}

func (j *JobBase) After(cfg *Config, jobID uuid.UUID, jobName string) {}

func (j *JobBase) AfterErr(cfg *Config, jobID uuid.UUID, jobName string, err error) {}

func (j *JobBase) Before(cfg *Config, jobID uuid.UUID, jobName string) {}

func (j *JobBase) Task(cfg *Config) error {
	return nil
}

// ----- Scheduler server
func NewSchedulerServer(cfg *Config, options ...gocron.SchedulerOption) (*SchedulerServer, error) {
	server, err := gocron.NewScheduler(options...)
	if err != nil {
		SchedulerLogger.Error(err.Error())
		return nil, err
	}

	return &SchedulerServer{
		Config: cfg,
		Server: server,
	}, nil

}

type SchedulerServer struct {
	Config *Config
	Server gocron.Scheduler
}

func (s *SchedulerServer) RegisterJob(job Job) error {
	options := make([]gocron.JobOption, 0)

	// setup job name
	options = append(options, gocron.WithName(job.Name()))

	// setup job event listener
	options = append(options, gocron.WithEventListeners(
		gocron.AfterJobRuns(func(jobID uuid.UUID, jobName string) {
			job.After(s.Config, jobID, jobName)
		}),
		gocron.AfterJobRunsWithError(func(jobID uuid.UUID, jobName string, err error) {
			job.AfterErr(s.Config, jobID, jobName, err)
		}),
		gocron.BeforeJobRuns(func(jobID uuid.UUID, jobName string) {
			job.Before(s.Config, jobID, jobName)
		}),
	))

	j, err := s.Server.NewJob(job.Duration(), gocron.NewTask(func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("%v", r)
			}
		}()
		err = job.Task(s.Config)
		return
	}), options...)
	if err != nil {
		SchedulerLogger.Error("failed run job", "name", job.Name())
		return err
	}

	SchedulerLogger.Info("start run job", "id", j.ID(), "name", j.Name())
	return nil
}

// ----- TODO
// 1. make auto create log run table
// 2. make default action before and and after job running
