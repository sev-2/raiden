package raiden_test

import (
	"os"
	"os/exec"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/sev-2/raiden"
	"github.com/stretchr/testify/assert"
)

func TestLog(t *testing.T) {
	raiden.SetLogLevel(hclog.Trace)

	raiden.Debug("Debug message")
	raiden.Info("Info message")
	raiden.Warning("Warning message")
	raiden.Error("Error message")

	if os.Getenv("FATAL") == "1" {
		raiden.Fatal("Fatal message")
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestLog")
	cmd.Env = append(os.Environ(), "FATAL=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func TestPanic(t *testing.T) {
	assert.Panics(t, func() {
		raiden.Panic("Panic message")
	})
}
