package roles_test

import (
	"fmt"
	"testing"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/resource/roles"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
)

func TestBuildMigrateData(t *testing.T) {
	extractedLocalData := state.ExtractRoleResult{
		New: []objects.Role{
			{Name: "role4"},
		},
		Existing: []objects.Role{
			{Name: "role2"},
			{Name: "role3"},
		},
		Delete: []objects.Role{
			{Name: "role1"},
		},
	}

	supabaseRoles := []objects.Role{
		{Name: "role1"},
		{Name: "role2"},
		{Name: "role3"},
	}

	migrateData, err := roles.BuildMigrateData(extractedLocalData, supabaseRoles)
	assert.NoError(t, err)
	assert.Equal(t, 4, len(migrateData))
}

func TestBuildMigrateItem(t *testing.T) {
	localRoles := []objects.Role{
		{Name: "role1"},
		{Name: "role2"},
	}

	supabaseRoles := []objects.Role{
		{Name: "role1"},
		{Name: "role3"},
	}

	migrateData, err := roles.BuildMigrateItem(supabaseRoles, localRoles)
	assert.NoError(t, err)
	fmt.Println(migrateData)
}

func TestMigrate(t *testing.T) {
	config := &raiden.Config{}
	stateChan := make(chan any)
	defer close(stateChan)

	migrateItems := []roles.MigrateItem{
		{
			Type:    "create",
			NewData: objects.Role{Name: "role1"},
		},
	}

	errors := roles.Migrate(config, migrateItems, stateChan, roles.ActionFunc)
	assert.Equal(t, 1, len(errors))
}
