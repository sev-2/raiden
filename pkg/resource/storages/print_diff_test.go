package storages_test

import (
	"bytes"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/sev-2/raiden/pkg/resource/migrator"
	"github.com/sev-2/raiden/pkg/resource/storages"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
)

// TestPrintDiffResult tests the PrintDiffResult function
func TestPrintDiffResult(t *testing.T) {
	diffResult := []storages.CompareDiffResult{
		{
			IsConflict:     true,
			SourceResource: objects.Bucket{Name: "source_bucket"},
			TargetResource: objects.Bucket{Name: "target_bucket"},
		},
	}

	err := storages.PrintDiffResult(diffResult)
	assert.EqualError(t, err, "canceled import process, you have conflict in storage. please fix it first")
}

// TestPrintDiff tests the PrintDiff function
func TestPrintDiff(t *testing.T) {

	if os.Getenv("TEST_RUN") == "1" {
		diffData := storages.CompareDiffResult{
			IsConflict:     true,
			SourceResource: objects.Bucket{Name: "source_bucket"},
			TargetResource: objects.Bucket{Name: "target_bucket"},
			DiffItems: objects.UpdateBucketParam{
				ChangeItems: []objects.UpdateBucketType{
					objects.UpdateBucketIsPublic,
					objects.UpdateBucketAllowedMimeTypes,
					objects.UpdateBucketFileSizeLimit,
				},
			},
		}

		storages.PrintDiff(diffData)
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
	items := []storages.MigrateItem{
		{
			Type:    migrator.MigrateTypeCreate,
			NewData: objects.Bucket{Name: "new_bucket"},
		},
		{
			Type:    migrator.MigrateTypeUpdate,
			NewData: objects.Bucket{Name: "update_bucket"},
		},
		{
			Type:    migrator.MigrateTypeDelete,
			OldData: objects.Bucket{Name: "delete_bucket"},
		},
	}

	diffMessage := storages.GetDiffChangeMessage(items)
	assert.Contains(t, diffMessage, "New Storage")
	assert.Contains(t, diffMessage, "Update Storage")
	assert.Contains(t, diffMessage, "Delete Storage")
}

// TestGenerateDiffMessage tests the GenerateDiffMessage function
func TestGenerateDiffMessage(t *testing.T) {
	diffType := storages.DiffTypeUpdate
	updateType := objects.UpdateBucketIsPublic
	value := "true"
	changeValue := "false"

	diffMessage, err := storages.GenerateDiffMessage("test_bucket", diffType, updateType, value, changeValue)
	assert.NoError(t, err)
	assert.Contains(t, diffMessage, "Public() bool")
}

// TestGenerateDiffChangeMessage tests the GenerateDiffChangeMessage function
func TestGenerateDiffChangeMessage(t *testing.T) {
	newData := []string{"new_bucket1", "new_bucket2"}
	updateData := []string{"update_bucket1", "update_bucket2"}
	deleteData := []string{"delete_bucket1", "delete_bucket2"}

	diffMessage, err := storages.GenerateDiffChangeMessage(newData, updateData, deleteData)
	assert.NoError(t, err)
	assert.Contains(t, diffMessage, "New Storage")
	assert.Contains(t, diffMessage, "Update Storage")
	assert.Contains(t, diffMessage, "Delete Storage")
}

// TestGenerateDiffChangeUpdateMessage tests the GenerateDiffChangeUpdateMessage function
func TestGenerateDiffChangeUpdateMessage(t *testing.T) {
	item := storages.MigrateItem{
		NewData: objects.Bucket{Name: "new_bucket"},
		OldData: objects.Bucket{Name: "old_bucket"},
		MigrationItems: objects.UpdateBucketParam{
			ChangeItems: []objects.UpdateBucketType{objects.UpdateBucketIsPublic},
		},
	}

	diffMessage, err := storages.GenerateDiffChangeUpdateMessage("test_bucket", item)
	assert.NoError(t, err)
	assert.Contains(t, diffMessage, "Update Storage test_bucket")
}
