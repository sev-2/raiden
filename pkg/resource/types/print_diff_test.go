package types_test

import (
	"bytes"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/sev-2/raiden/pkg/resource/types"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
)

// TestPrintDiffResult tests the PrintDiffResult type
func TestPrintDiffResult(t *testing.T) {
	diffResult := []types.CompareDiffResult{
		{
			IsConflict:     true,
			SourceResource: objects.Type{Name: "source_type"},
			TargetResource: objects.Type{Name: "target_type"},
		},
	}

	err := types.PrintDiffResult(diffResult)
	assert.EqualError(t, err, "canceled import process, you have conflict in type. please fix it first")
}

// TestPrintDiff tests the PrintDiff type
func TestPrintDiff(t *testing.T) {

	if os.Getenv("TEST_RUN") == "1" {
		diffData := types.CompareDiffResult{
			IsConflict:     true,
			SourceResource: objects.Type{Name: "source_type"},
			TargetResource: objects.Type{Name: "target_type"},
		}

		types.PrintDiff(diffData)
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

	migratedItems := objects.UpdateTypeParam{
		OldData: objects.Type{Name: "source_type"},
		ChangeItems: []objects.UpdateDataType{
			objects.UpdateTypeName,
			objects.UpdateTypeAttributes,
			objects.UpdateTypeComment,
			objects.UpdateTypeEnums,
			objects.UpdateTypeFormat,
			objects.UpdateTypeSchema,
		},
	}

	successDiffData := types.CompareDiffResult{
		IsConflict: false,
		SourceResource: objects.Type{
			Name:       "source_type",
			Schema:     "public",
			Format:     "",
			Enums:      []string{"test_1", "test_2"},
			Attributes: []string{},
			Comment:    nil,
		},
		TargetResource: objects.Type{
			Name:       "target_type",
			Schema:     "test",
			Format:     "",
			Enums:      []string{"test_1", "test_3"},
			Attributes: []string{},
			Comment:    nil,
		},
		DiffItems: migratedItems,
	}

	types.PrintDiff(successDiffData)
}

// TestGetDiffChangeMessage tests the GetDiffChangeMessage type
// func TestGetDiffChangeMessage(t *testing.T) {
// 	items := []types.MigrateItem{
// 		{
// 			Type:    migrator.MigrateTypeCreate,
// 			NewData: objects.Type{Name: "new_type"},
// 		},
// 		{
// 			Type: migrator.MigrateTypeUpdate,
// 			NewData: objects.Type{
// 				Name:              "source_type",
// 				CanBypassRLS:      true,
// 				CanCreateDB:       true,
// 				CanCreateRole:     true,
// 				CanLogin:          true,
// 				Config:            map[string]interface{}{"key": "value"},
// 				ConnectionLimit:   10,
// 				InheritRole:       true,
// 				IsReplicationRole: true,
// 				IsSuperuser:       true,
// 				ValidUntil:        &objects.SupabaseTime{},
// 			},
// 			OldData: objects.Type{
// 				Name:              "target_type",
// 				CanBypassRLS:      false,
// 				CanCreateDB:       false,
// 				CanCreateRole:     false,
// 				CanLogin:          false,
// 				Config:            map[string]interface{}{"key": "new-value"},
// 				ConnectionLimit:   20,
// 				InheritRole:       false,
// 				IsReplicationRole: false,
// 				IsSuperuser:       false,
// 				ValidUntil:        &objects.SupabaseTime{},
// 			},
// 			MigrationItems: objects.UpdateDataType{
// 				OldData: objects.Type{Name: "source_type"},
// 				ChangeItems: []objects.UpdateRoleType{
// 					objects.UpdateRoleName,
// 					objects.UpdateRoleCanBypassRls,
// 					objects.UpdateRoleCanCreateDb,
// 					objects.UpdateRoleCanCreateRole,
// 					objects.UpdateRoleCanLogin,
// 					objects.UpdateRoleConfig,
// 					objects.UpdateConnectionLimit,
// 					objects.UpdateRoleInheritRole,
// 					objects.UpdateRoleIsReplication,
// 					objects.UpdateRoleIsSuperUser,
// 					objects.UpdateRoleValidUntil,
// 				},
// 			},
// 		},
// 		{
// 			Type:    migrator.MigrateTypeDelete,
// 			OldData: objects.Type{Name: "delete_type"},
// 		},
// 	}

// 	diffMessage := types.GetDiffChangeMessage(items)
// 	assert.Contains(t, diffMessage, "New Type"
// 	assert.Contains(t, diffMessage, "Update Type")
// 	assert.Contains(t, diffMessage, "Delete Type")
// }

// TestGenerateDiffMessage tests the GenerateDiffMessage type
func TestGenerateDiffMessage(t *testing.T) {
	diffType := types.DiffTypeUpdate
	updateType := objects.UpdateTypeEnums
	value := "true"
	changeValue := "false"

	diffMessage, err := types.GenerateDiffMessage("test_type", diffType, updateType, value, changeValue)
	assert.NoError(t, err)
	assert.Contains(t, diffMessage, "Enums")
}

// TestGenerateDiffChangeMessage tests the GenerateDiffChangeMessage type
func TestGenerateDiffChangeMessage(t *testing.T) {
	newData := []string{"new_type1", "new_type2"}
	updateData := []string{"update_type1", "update_type2"}
	deleteData := []string{"delete_type1", "delete_type2"}

	diffMessage, err := types.GenerateDiffChangeMessage(newData, updateData, deleteData)
	assert.NoError(t, err)
	assert.Contains(t, diffMessage, "New Type")
	assert.Contains(t, diffMessage, "Update Type")
	assert.Contains(t, diffMessage, "Delete Type")
}
