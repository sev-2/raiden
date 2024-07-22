package rpc_test

import (
	"bytes"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/sev-2/raiden/pkg/resource/migrator"
	"github.com/sev-2/raiden/pkg/resource/rpc"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
)

// TestPrintDiffResult tests the PrintDiffResult function
func TestPrintDiffResult(t *testing.T) {
	diffResult := []rpc.CompareDiffResult{
		{
			IsConflict:     true,
			SourceResource: objects.Function{Name: "source_function"},
			TargetResource: objects.Function{Name: "target_function"},
		},
	}

	err := rpc.PrintDiffResult(diffResult)
	assert.EqualError(t, err, "canceled import process, you have conflict rpc function. please fix it first")
}

// TestPrintDiff tests the PrintDiff function
func TestPrintDiff(t *testing.T) {

	if os.Getenv("TEST_RUN") == "1" {
		diffData := rpc.CompareDiffResult{
			IsConflict:     true,
			SourceResource: objects.Function{Name: "source_function"},
			TargetResource: objects.Function{Name: "target_function"},
		}

		rpc.PrintDiff(diffData)
		return
	}

	var outb, errb bytes.Buffer
	cmd := exec.Command(os.Args[0], "-test.run=TestPrintDiff")
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	cmd.Env = append(os.Environ(), "TEST_RUN=1")
	err := cmd.Start()
	assert.NoError(t, err)

	time.Sleep(1 * time.Second)
	err1 := cmd.Process.Signal(syscall.SIGTERM)
	assert.NoError(t, err1)

	assert.Contains(t, outb.String(), "Found diff")
	assert.Contains(t, outb.String(), "End found diff")
}

// TestGetDiffChangeMessage tests the GetDiffChangeMessage function
func TestGetDiffChangeMessage(t *testing.T) {
	items := []rpc.MigrateItem{
		{
			Type:    migrator.MigrateTypeCreate,
			NewData: objects.Function{Name: "new_function"},
		},
		{
			Type:    migrator.MigrateTypeUpdate,
			NewData: objects.Function{Name: "update_function"},
		},
		{
			Type:    migrator.MigrateTypeDelete,
			OldData: objects.Function{Name: "delete_function"},
		},
	}

	diffMessage := rpc.GetDiffChangeMessage(items)
	assert.Contains(t, diffMessage, "New Rpc")
	assert.Contains(t, diffMessage, "Update Rpc")
	assert.Contains(t, diffMessage, "Delete Rpc")
}

// TestGenerateDiffMessage tests the GenerateDiffMessage function
func TestGenerateDiffMessage(t *testing.T) {
	value := "true"
	changeValue := "false"

	diffMessage, err := rpc.GenerateDiffMessage("test_function", value, changeValue)
	assert.NoError(t, err)
	assert.Contains(t, diffMessage, "$function")
}

// TestGenerateDiffChangeMessage tests the GenerateDiffChangeMessage function
func TestGenerateDiffChangeMessage(t *testing.T) {
	newData := []string{"new_function1", "new_function2"}
	updateData := []string{"update_function1", "update_function2"}
	deleteData := []string{"delete_function1", "delete_function2"}

	diffMessage, err := rpc.GenerateDiffChangeMessage(newData, updateData, deleteData)
	assert.NoError(t, err)
	assert.Contains(t, diffMessage, "New Rpc")
	assert.Contains(t, diffMessage, "Update Rpc")
	assert.Contains(t, diffMessage, "Delete Rpc")
}
