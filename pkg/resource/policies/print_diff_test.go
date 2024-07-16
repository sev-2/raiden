package policies_test

import (
	"testing"

	"github.com/sev-2/raiden/pkg/resource/migrator"
	"github.com/sev-2/raiden/pkg/resource/policies"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
)

func TestGetDiffChangeMessage(t *testing.T) {
	items := []policies.MigrateItem{
		{
			Type:    migrator.MigrateTypeCreate,
			NewData: objects.Policy{Name: "Policy1"},
		},
		{
			Type:           migrator.MigrateTypeUpdate,
			NewData:        objects.Policy{Name: "Policy2", Definition: "new_def"},
			OldData:        objects.Policy{Name: "Policy2", Definition: "old_def"},
			MigrationItems: objects.UpdatePolicyParam{ChangeItems: []objects.UpdatePolicyType{objects.UpdatePolicyDefinition}},
		},
		{
			Type:    migrator.MigrateTypeDelete,
			OldData: objects.Policy{Name: "Policy3"},
		},
	}

	diffMessage := policies.GetDiffChangeMessage(items)
	assert.Contains(t, diffMessage, "New Policy")
	assert.Contains(t, diffMessage, "- Policy1")
	assert.Contains(t, diffMessage, "Update Policy")
	assert.Contains(t, diffMessage, "- Policy1")
	assert.Contains(t, diffMessage, "Change Configuration")
	assert.Contains(t, diffMessage, "- definition : old_def >>> new_def")
	assert.Contains(t, diffMessage, "Delete Policy")
	assert.Contains(t, diffMessage, "- Policy3")
}

func TestGenerateDiffChangeMessage(t *testing.T) {
	newData := []string{"- Policy1"}
	updateData := []string{"- Policy2"}
	deleteData := []string{"- Policy3"}

	diffMessage, err := policies.GenerateDiffChangeMessage(newData, updateData, deleteData)
	assert.NoError(t, err)
	assert.Contains(t, diffMessage, "New Policy")
	assert.Contains(t, diffMessage, "- Policy1")
	assert.Contains(t, diffMessage, "Update Policy")
	assert.Contains(t, diffMessage, "- Policy2")
	assert.Contains(t, diffMessage, "Delete Policy")
	assert.Contains(t, diffMessage, "- Policy3")
}

func TestGenerateDiffChangeUpdateMessage(t *testing.T) {
	oldCheck := "old_check"
	newCheck := "new_check"

	item := policies.MigrateItem{
		NewData: objects.Policy{Name: "Policy2", Definition: "new_def", Check: &newCheck},
		OldData: objects.Policy{Name: "Policy2", Definition: "old_def", Check: &oldCheck},
		MigrationItems: objects.UpdatePolicyParam{
			ChangeItems: []objects.UpdatePolicyType{
				objects.UpdatePolicyName,
				objects.UpdatePolicyDefinition,
				objects.UpdatePolicyCheck,
			},
		},
	}

	diffMessage, err := policies.GenerateDiffChangeUpdateMessage("Policy2", item)
	assert.NoError(t, err)
	assert.Contains(t, diffMessage, "Update Policy")
	assert.Contains(t, diffMessage, "- name : Policy2 >>> Policy2")
	assert.Contains(t, diffMessage, "- definition : old_def >>> new_def")
	assert.Contains(t, diffMessage, "- check : old_check >>> new_check")
}
