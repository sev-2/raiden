package policies_test

import (
	"testing"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/resource/migrator"
	"github.com/sev-2/raiden/pkg/resource/policies"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
)

func TestBuildMigrateData(t *testing.T) {
	extractedLocalData := state.ExtractedPolicies{
		New: []objects.Policy{
			{Name: "Policy1", Definition: "def1", Roles: []string{"role1"}},
		},
		Existing: []objects.Policy{
			{Name: "Policy2", Definition: "def2", Roles: []string{"role2"}},
		},
		Delete: []objects.Policy{
			{Name: "Policy3", Definition: "def3", Roles: []string{"role3"}},
		},
	}

	supabaseData := []objects.Policy{
		{Name: "Policy2", Definition: "def2", Roles: []string{"role2"}},
		{Name: "Policy4", Definition: "def4", Roles: []string{"role4"}},
	}

	migrateData, err := policies.BuildMigrateData(extractedLocalData, supabaseData)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(migrateData))

	assert.Equal(t, migrator.MigrateTypeIgnore, migrateData[0].Type)
	assert.Equal(t, "Policy2", migrateData[0].NewData.Name)

	assert.Equal(t, migrator.MigrateTypeCreate, migrateData[1].Type)
	assert.Equal(t, "Policy1", migrateData[1].NewData.Name)
}

func TestBuildMigrateItem(t *testing.T) {
	localData := []objects.Policy{
		{Name: "Policy1", Definition: "def1", Roles: []string{"role1"}},
		{Name: "Policy2", Definition: "def2", Roles: []string{"role2"}},
	}

	supabaseData := []objects.Policy{
		{Name: "Policy1", Definition: "def1", Roles: []string{"role1"}},
		{Name: "Policy2", Definition: "diff_def2", Roles: []string{"role2"}},
	}

	migrateData, err := policies.BuildMigrateItem(supabaseData, localData)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(migrateData))
	assert.Equal(t, migrator.MigrateTypeIgnore, migrateData[0].Type)
	assert.Equal(t, "Policy1", migrateData[0].NewData.Name)
}

func TestMigrate(t *testing.T) {
	config := &raiden.Config{}
	stateChan := make(chan any)
	defer close(stateChan)

	migrateItems := []policies.MigrateItem{
		{
			Type:    migrator.MigrateTypeCreate,
			NewData: objects.Policy{Name: "Policy1"},
		},
		{
			Type:    migrator.MigrateTypeUpdate,
			NewData: objects.Policy{Name: "Policy2"},
		},
		{
			Type:    migrator.MigrateTypeDelete,
			OldData: objects.Policy{Name: "Policy3"},
		},
	}

	actions := policies.ActionFunc
	errors := policies.Migrate(config, migrateItems, stateChan, actions)

	assert.Equal(t, 3, len(errors))
}
