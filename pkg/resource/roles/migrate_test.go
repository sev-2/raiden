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

func TestBuildMigrateData_InheritRoleItems(t *testing.T) {
	child := objects.Role{
		Name:        "child_role",
		InheritRole: true,
		InheritRoles: []*objects.Role{
			{Name: "parent_role"},
			{Name: "Parent_Role"},
			{Name: ""},
			nil,
		},
	}

	flatData := state.ExtractRoleResult{New: []objects.Role{child}}

	migrateData, err := roles.BuildMigrateData(flatData, nil)
	assert.NoError(t, err)

	if assert.Len(t, migrateData, 1) {
		item := migrateData[0]
		assert.Equal(t, migrator.MigrateTypeCreate, item.Type)
		if assert.Len(t, item.MigrationItems.ChangeInheritItems, 1) {
			inheritItem := item.MigrationItems.ChangeInheritItems[0]
			assert.Equal(t, objects.UpdateRoleInheritGrant, inheritItem.Type)
			assert.Equal(t, "parent_role", inheritItem.Role.Name)
		}
	}
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

func TestBuildMigrateItem_InheritDiff(t *testing.T) {
	supabaseRoles := []objects.Role{
		{ID: 1, Name: "child_role", InheritRole: true, InheritRoles: []*objects.Role{{Name: "legacy_parent"}}},
		{ID: 2, Name: "legacy_parent"},
	}

	localRoles := []objects.Role{
		{ID: 1, Name: "child_role", InheritRole: true, InheritRoles: []*objects.Role{{Name: "new_parent"}}}, // Same ID to match with supabase child_role
		{ID: 3, Name: "new_parent"},
	}

	migrateData, err := roles.BuildMigrateItem(supabaseRoles, localRoles)
	assert.NoError(t, err)

	var updateItem *roles.MigrateItem
	for i := range migrateData {
		if migrateData[i].Type == migrator.MigrateTypeUpdate {
			updateItem = &migrateData[i]
			break
		}
	}

	if assert.NotNil(t, updateItem) {
		param := updateItem.MigrationItems
		if assert.Len(t, param.ChangeInheritItems, 2) {
			actions := map[objects.UpdateRoleInheritType]string{}
			for _, item := range param.ChangeInheritItems {
				actions[item.Type] = item.Role.Name
			}
			assert.Equal(t, "legacy_parent", actions[objects.UpdateRoleInheritRevoke])
			assert.Equal(t, "new_parent", actions[objects.UpdateRoleInheritGrant])
		}
	}
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

func TestMigrateRole_UpdateAndDelete(t *testing.T) {
	cfg := &raiden.Config{}
	stateChan := make(chan any, 2)

	// Test update path
	updateCalled := false
	deleteCalled := false
	actions := roles.MigrateActionFunc{
		CreateFunc: func(_ *raiden.Config, role objects.Role) (objects.Role, error) {
			return role, nil
		},
		UpdateFunc: func(_ *raiden.Config, role objects.Role, param objects.UpdateRoleParam) error {
			updateCalled = true
			return nil
		},
		DeleteFunc: func(_ *raiden.Config, role objects.Role) error {
			deleteCalled = true
			return nil
		},
	}

	// Test with Update type
	migrateItems := []roles.MigrateItem{
		{
			Type:    migrator.MigrateTypeUpdate,
			NewData: objects.Role{Name: "test_role"},
		},
	}

	errs := roles.Migrate(cfg, migrateItems, stateChan, actions)
	assert.Empty(t, errs)
	assert.True(t, updateCalled)

	// Test with Delete type - reset the flag
	updateCalled = false
	migrateItemsDelete := []roles.MigrateItem{
		{
			Type:    migrator.MigrateTypeDelete,
			OldData: objects.Role{Name: "delete_role"},
		},
	}

	errs2 := roles.Migrate(cfg, migrateItemsDelete, stateChan, actions)
	assert.Empty(t, errs2)
	assert.True(t, deleteCalled)
}

func TestMigrate_InvalidType(t *testing.T) {
	cfg := &raiden.Config{}
	stateChan := make(chan any)

	actions := roles.MigrateActionFunc{
		CreateFunc: func(_ *raiden.Config, role objects.Role) (objects.Role, error) {
			return role, nil
		},
		UpdateFunc: func(_ *raiden.Config, role objects.Role, param objects.UpdateRoleParam) error {
			return nil
		},
		DeleteFunc: func(_ *raiden.Config, role objects.Role) error {
			return nil
		},
	}

	// Test with unknown type - should call the default case in migrateRole
	migrateItems := []roles.MigrateItem{
		{
			Type:    "invalid_type", // Not a valid type
			NewData: objects.Role{Name: "test_role"},
		},
	}

	errs := roles.Migrate(cfg, migrateItems, stateChan, actions)
	assert.Empty(t, errs) // Should not have errors since default case returns nil
}
