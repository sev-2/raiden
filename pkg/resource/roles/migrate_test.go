package roles_test

import (
	"fmt"
	"testing"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/resource/migrator"
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

func TestMigrate_CreateRoleWithInheritance(t *testing.T) {
	cfg := &raiden.Config{}
	stateChan := make(chan any, 1)

	var createdRoles []objects.Role
	var updateParams []objects.UpdateRoleParam

	actions := roles.MigrateActionFunc{
		CreateFunc: func(_ *raiden.Config, role objects.Role) (objects.Role, error) {
			createdRoles = append(createdRoles, role)
			role.ID = 99
			return role, nil
		},
		UpdateFunc: func(_ *raiden.Config, role objects.Role, param objects.UpdateRoleParam) error {
			updateParams = append(updateParams, param)
			return nil
		},
		DeleteFunc: func(*raiden.Config, objects.Role) error { return nil },
	}

	migrateItems := []roles.MigrateItem{
		{
			Type: migrator.MigrateTypeCreate,
			NewData: objects.Role{
				Name:        "child_role",
				InheritRole: true,
				InheritRoles: []*objects.Role{
					{Name: "parent_role"},
					{Name: "Parent_Role"}, // duplicate with different casing should be deduplicated
				},
			},
		},
	}

	errs := roles.Migrate(cfg, migrateItems, stateChan, actions)
	assert.Empty(t, errs)
	assert.Len(t, createdRoles, 1)

	if assert.Len(t, updateParams, 1) {
		param := updateParams[0]
		assert.Equal(t, 99, param.OldData.ID)
		assert.Equal(t, "child_role", param.OldData.Name)
		assert.True(t, param.OldData.InheritRole)
		if assert.Len(t, param.ChangeInheritItems, 1) {
			assert.Equal(t, objects.UpdateRoleInheritGrant, param.ChangeInheritItems[0].Type)
			assert.Equal(t, "parent_role", param.ChangeInheritItems[0].Role.Name)
		}
	}

	select {
	case v := <-stateChan:
		roleItem, ok := v.(*roles.MigrateItem)
		if assert.True(t, ok) {
			assert.Equal(t, "child_role", roleItem.NewData.Name)
			assert.Len(t, roleItem.NewData.InheritRoles, 1)
			if assert.NotNil(t, roleItem.NewData.InheritRoles[0]) {
				assert.Equal(t, "parent_role", roleItem.NewData.InheritRoles[0].Name)
			}
		}
	default:
		t.Fatal("expected state update from migrate")
	}
}
