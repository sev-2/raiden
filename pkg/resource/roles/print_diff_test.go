package roles_test

import (
	"bytes"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/sev-2/raiden/pkg/resource/migrator"
	"github.com/sev-2/raiden/pkg/resource/roles"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
)

// TestPrintDiffResult tests the PrintDiffResult role
func TestPrintDiffResult(t *testing.T) {
	diffResult := []roles.CompareDiffResult{
		{
			IsConflict:     true,
			SourceResource: objects.Role{Name: "source_role"},
			TargetResource: objects.Role{Name: "target_role"},
		},
	}

	err := roles.PrintDiffResult(diffResult)
	assert.EqualError(t, err, "canceled import process, you have conflict in role. please fix it first")
}

// TestPrintDiff tests the PrintDiff role
func TestPrintDiff(t *testing.T) {

	if os.Getenv("TEST_RUN") == "1" {
		diffData := roles.CompareDiffResult{
			IsConflict:     true,
			SourceResource: objects.Role{Name: "source_role"},
			TargetResource: objects.Role{Name: "target_role"},
		}

		roles.PrintDiff(diffData)
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

	assert.Contains(t, outb.String(), "PASS")
}

// TestGetDiffChangeMessage tests the GetDiffChangeMessage role
func TestGetDiffChangeMessage(t *testing.T) {
	items := []roles.MigrateItem{
		{
			Type:    migrator.MigrateTypeCreate,
			NewData: objects.Role{Name: "new_role"},
		},
		{
			Type:    migrator.MigrateTypeUpdate,
			NewData: objects.Role{Name: "update_role"},
		},
		{
			Type:    migrator.MigrateTypeDelete,
			OldData: objects.Role{Name: "delete_role"},
		},
	}

	diffMessage := roles.GetDiffChangeMessage(items)
	assert.Contains(t, diffMessage, "New Role")
	assert.Contains(t, diffMessage, "Update Role")
	assert.Contains(t, diffMessage, "Delete Role")
}

// TestGenerateDiffMessage tests the GenerateDiffMessage role
func TestGenerateDiffMessage(t *testing.T) {
	diffType := roles.DiffTypeUpdate
	updateType := objects.UpdateRoleCanBypassRls
	value := "true"
	changeValue := "false"

	diffMessage, err := roles.GenerateDiffMessage("test_role", diffType, updateType, value, changeValue)
	assert.NoError(t, err)
	assert.Contains(t, diffMessage, "CanBypassRls")
}

// TestGenerateDiffChangeMessage tests the GenerateDiffChangeMessage role
func TestGenerateDiffChangeMessage(t *testing.T) {
	newData := []string{"new_role1", "new_role2"}
	updateData := []string{"update_role1", "update_role2"}
	deleteData := []string{"delete_role1", "delete_role2"}

	diffMessage, err := roles.GenerateDiffChangeMessage(newData, updateData, deleteData)
	assert.NoError(t, err)
	assert.Contains(t, diffMessage, "New Role")
	assert.Contains(t, diffMessage, "Update Role")
	assert.Contains(t, diffMessage, "Delete Role")
}