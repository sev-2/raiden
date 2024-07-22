package raiden_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/sev-2/raiden"
	"github.com/stretchr/testify/assert"
)

type SampleJob struct {
	raiden.JobBase
}

func (j *SampleJob) Name() string {
	return "SampleJob"
}

func (j *SampleJob) Duration() raiden.JobDuration {
	return gocron.DurationRandomJob(4*time.Minute, 6*time.Minute)
}

func TestScheduler_SetTracer(t *testing.T) {
	conf := loadConfig()
	ss, err := raiden.NewSchedulerServer(conf, gocron.WithLimitConcurrentJobs(2, gocron.LimitModeReschedule))
	assert.NoError(t, err)
	ss.SetTracer(nil)
}

func TestScheduler_RegisterJob(t *testing.T) {
	conf := loadConfig()
	ss, err := raiden.NewSchedulerServer(conf, gocron.WithLimitConcurrentJobs(2, gocron.LimitModeReschedule))
	assert.NoError(t, err)

	err1 := ss.RegisterJob(&SampleJob{})
	assert.NoError(t, err1)
}

func TestScheduler_Start(t *testing.T) {
	conf := loadConfig()
	ss, err := raiden.NewSchedulerServer(conf, gocron.WithLimitConcurrentJobs(2, gocron.LimitModeReschedule))
	assert.NoError(t, err)
	ss.Start()

	err1 := ss.Stop(context.Background())
	assert.NoError(t, err1)
}
