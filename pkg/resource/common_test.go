package resource_test

import (
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/resource"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestFlags_All(t *testing.T) {
	flags := &resource.Flags{}
	assert.True(t, flags.All())

	flags.RpcOnly = true
	assert.False(t, flags.All())

	flags.RpcOnly = false
	flags.RolesOnly = true
	assert.False(t, flags.All())

	flags.RolesOnly = false
	flags.ModelsOnly = true
	assert.False(t, flags.All())

	flags.ModelsOnly = false
	flags.StoragesOnly = true
	assert.False(t, flags.All())
}

func TestFlags_BindLog(t *testing.T) {
	cmd := &cobra.Command{}
	flags := &resource.Flags{}
	flags.BindLog(cmd)

	debugFlag := cmd.PersistentFlags().Lookup("debug")
	assert.NotNil(t, debugFlag)
	assert.False(t, flags.DebugMode)

	traceFlag := cmd.PersistentFlags().Lookup("trace")
	assert.NotNil(t, traceFlag)
	assert.False(t, flags.TraceMode)
}

func TestFlags_CheckAndActivateDebug(t *testing.T) {
	cmd := &cobra.Command{}
	flags := resource.Flags{
		DebugMode: true,
		TraceMode: false,
	}
	flags.CheckAndActivateDebug(cmd)
	assert.Equal(t, hclog.Debug, logger.HcLog().GetLevel())

	flags.DebugMode = false
	flags.TraceMode = true
	flags.CheckAndActivateDebug(cmd)
	assert.Equal(t, hclog.Trace, logger.HcLog().GetLevel())
}

func TestPreRun(t *testing.T) {
	if os.Getenv("TEST_RUN") == "1" {
		err := resource.PreRun("valid_path")
		assert.NoError(t, err)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestPreRun")
	cmd.Env = append(os.Environ(), "TEST_RUN=1")
	err := cmd.Start()
	assert.NoError(t, err)

	time.Sleep(1 * time.Second)
	err1 := cmd.Process.Signal(syscall.SIGTERM)
	assert.NoError(t, err1)
}
