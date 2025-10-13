package migrator_test

import (
	"errors"
	"testing"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/resource/migrator"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
)

func TestMigrateTypeConstants(t *testing.T) {
	assert.Equal(t, migrator.MigrateType("ignore"), migrator.MigrateTypeIgnore)
	assert.Equal(t, migrator.MigrateType("create"), migrator.MigrateTypeCreate)
	assert.Equal(t, migrator.MigrateType("update"), migrator.MigrateTypeUpdate)
	assert.Equal(t, migrator.MigrateType("delete"), migrator.MigrateTypeDelete)
}

func TestMigrateResource(t *testing.T) {
	cfg := &raiden.Config{}
	stateChan := make(chan any, 10)
	
	// Test successful migration
	resources := []migrator.MigrateItem[string, string]{
		{
			Type:    migrator.MigrateTypeCreate,
			NewData: "test-resource",
		},
	}

	createFunc := func(cfg *raiden.Config, param string) (string, error) {
		return "created-" + param, nil
	}
	
	updateFunc := func(cfg *raiden.Config, param string, items string) error {
		return nil
	}
	
	deleteFunc := func(cfg *raiden.Config, param string) error {
		return nil
	}
	
	actionFuncs := migrator.MigrateActionFunc[string, string]{
		CreateFunc: createFunc,
		UpdateFunc: updateFunc,
		DeleteFunc: deleteFunc,
	}

	// Test with default migrator
	errs := migrator.MigrateResource(cfg, resources, stateChan, actionFuncs, migrator.DefaultMigrator[string, string])
	assert.Empty(t, errs)
	
	// Test with error scenario
	createFuncWithError := func(cfg *raiden.Config, param string) (string, error) {
		return "", errors.New("creation failed")
	}
	
	actionFuncsWithError := migrator.MigrateActionFunc[string, string]{
		CreateFunc: createFuncWithError,
		UpdateFunc: updateFunc,
		DeleteFunc: deleteFunc,
	}
	
	errs2 := migrator.MigrateResource(cfg, resources, stateChan, actionFuncsWithError, migrator.DefaultMigrator[string, string])
	assert.NotEmpty(t, errs2)
}

func TestMigrateResource_WithAllTypes(t *testing.T) {
	cfg := &raiden.Config{}
	stateChan := make(chan any, 10)
	
	resources := []migrator.MigrateItem[string, string]{
		{
			Type:    migrator.MigrateTypeCreate,
			NewData: "create-resource",
		},
		{
			Type:    migrator.MigrateTypeUpdate,
			NewData: "update-resource",
			OldData: "old-resource",
		},
		{
			Type:    migrator.MigrateTypeDelete,
			OldData: "delete-resource",
		},
		{
			Type:    migrator.MigrateTypeIgnore,
			NewData: "ignore-resource",
		},
	}

	createFunc := func(cfg *raiden.Config, param string) (string, error) {
		return "created-" + param, nil
	}
	
	updateFunc := func(cfg *raiden.Config, param string, items string) error {
		return nil
	}
	
	deleteFunc := func(cfg *raiden.Config, param string) error {
		return nil
	}
	
	actionFuncs := migrator.MigrateActionFunc[string, string]{
		CreateFunc: createFunc,
		UpdateFunc: updateFunc,
		DeleteFunc: deleteFunc,
	}

	errs := migrator.MigrateResource(cfg, resources, stateChan, actionFuncs, migrator.DefaultMigrator[string, string])
	assert.Empty(t, errs)
}

func TestDefaultMigrator(t *testing.T) {
	cfg := &raiden.Config{}
	stateChan := make(chan any, 10)
	
	// Test create type
	createItem := migrator.MigrateItem[string, string]{
		Type:    migrator.MigrateTypeCreate,
		NewData: "test-create",
	}
	
	createFunc := func(cfg *raiden.Config, param string) (string, error) {
		return "created-" + param, nil
	}
	
	updateFunc := func(cfg *raiden.Config, param string, items string) error {
		return nil
	}
	
	deleteFunc := func(cfg *raiden.Config, param string) error {
		return nil
	}
	
	actionFuncs := migrator.MigrateActionFunc[string, string]{
		CreateFunc: createFunc,
		UpdateFunc: updateFunc,
		DeleteFunc: deleteFunc,
	}
	
	params := migrator.MigrateFuncParam[string, string]{
		Config:      cfg,
		StateChan:   stateChan,
		Data:        createItem,
		ActionFuncs: actionFuncs,
	}
	
	err := migrator.DefaultMigrator(params)
	assert.NoError(t, err)
	
	// Test update type
	updateItem := migrator.MigrateItem[string, string]{
		Type:    migrator.MigrateTypeUpdate,
		NewData: "test-update",
	}
	
	paramsUpdate := migrator.MigrateFuncParam[string, string]{
		Config:      cfg,
		StateChan:   stateChan,
		Data:        updateItem,
		ActionFuncs: actionFuncs,
	}
	
	err = migrator.DefaultMigrator(paramsUpdate)
	assert.NoError(t, err)
	
	// Test delete type
	deleteItem := migrator.MigrateItem[string, string]{
		Type:    migrator.MigrateTypeDelete,
		OldData: "test-delete",
	}
	
	paramsDelete := migrator.MigrateFuncParam[string, string]{
		Config:      cfg,
		StateChan:   stateChan,
		Data:        deleteItem,
		ActionFuncs: actionFuncs,
	}
	
	err = migrator.DefaultMigrator(paramsDelete)
	assert.NoError(t, err)
	
	// Test with error in create
	createFuncWithError := func(cfg *raiden.Config, param string) (string, error) {
		return "", errors.New("create failed")
	}
	
	actionFuncsWithError := migrator.MigrateActionFunc[string, string]{
		CreateFunc: createFuncWithError,
		UpdateFunc: updateFunc,
		DeleteFunc: deleteFunc,
	}
	
	paramsError := migrator.MigrateFuncParam[string, string]{
		Config:      cfg,
		StateChan:   stateChan,
		Data:        createItem,
		ActionFuncs: actionFuncsWithError,
	}
	
	err = migrator.DefaultMigrator(paramsError)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "create failed")
	
	// Test with error in update
	updateFuncWithError := func(cfg *raiden.Config, param string, items string) error {
		return errors.New("update failed")
	}
	
	actionFuncsUpdateError := migrator.MigrateActionFunc[string, string]{
		CreateFunc: createFunc,
		UpdateFunc: updateFuncWithError,
		DeleteFunc: deleteFunc,
	}
	
	paramsUpdateError := migrator.MigrateFuncParam[string, string]{
		Config:      cfg,
		StateChan:   stateChan,
		Data:        updateItem,
		ActionFuncs: actionFuncsUpdateError,
	}
	
	err = migrator.DefaultMigrator(paramsUpdateError)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update failed")
	
	// Test with error in delete
	deleteFuncWithError := func(cfg *raiden.Config, param string) error {
		return errors.New("delete failed")
	}
	
	actionFuncsDeleteError := migrator.MigrateActionFunc[string, string]{
		CreateFunc: createFunc,
		UpdateFunc: updateFunc,
		DeleteFunc: deleteFuncWithError,
	}
	
	paramsDeleteError := migrator.MigrateFuncParam[string, string]{
		Config:      cfg,
		StateChan:   stateChan,
		Data:        deleteItem,
		ActionFuncs: actionFuncsDeleteError,
	}
	
	err = migrator.DefaultMigrator(paramsDeleteError)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delete failed")
	
	// Test with unknown type (should return nil)
	unknownItem := migrator.MigrateItem[string, string]{
		Type:    "unknown",
		NewData: "test-unknown",
	}
	
	paramsUnknown := migrator.MigrateFuncParam[string, string]{
		Config:      cfg,
		StateChan:   stateChan,
		Data:        unknownItem,
		ActionFuncs: actionFuncs,
	}
	
	err = migrator.DefaultMigrator(paramsUnknown)
	assert.NoError(t, err)
}

func TestMigratePolicy(t *testing.T) {
	cfg := &raiden.Config{}
	stateChan := make(chan any, 10)
	
	policies := []migrator.MigrateItem[objects.Policy, objects.UpdatePolicyParam]{
		{
			Type: migrator.MigrateTypeCreate,
			NewData: objects.Policy{
				Name:  "test-policy-1",
				Table: "test-table-1",
			},
		},
		{
			Type: migrator.MigrateTypeUpdate,
			NewData: objects.Policy{
				Name:  "test-policy-2",
				Table: "test-table-1", // Same table as first policy
			},
		},
		{
			Type: migrator.MigrateTypeDelete,
			OldData: objects.Policy{
				Name:  "test-policy-3",
				Table: "test-table-2", // Different table
			},
		},
	}

	createFunc := func(cfg *raiden.Config, param objects.Policy) (objects.Policy, error) {
		return param, nil
	}
	
	updateFunc := func(cfg *raiden.Config, param objects.Policy, items objects.UpdatePolicyParam) error {
		return nil
	}
	
	deleteFunc := func(cfg *raiden.Config, param objects.Policy) error {
		return nil
	}
	
	actionFuncs := migrator.MigrateActionFunc[objects.Policy, objects.UpdatePolicyParam]{
		CreateFunc: createFunc,
		UpdateFunc: updateFunc,
		DeleteFunc: deleteFunc,
	}

	errs := migrator.MigratePolicy(cfg, policies, stateChan, actionFuncs)
	assert.Empty(t, errs)
}

func TestMigratePolicy_WithError(t *testing.T) {
	cfg := &raiden.Config{}
	stateChan := make(chan any, 10)
	
	policies := []migrator.MigrateItem[objects.Policy, objects.UpdatePolicyParam]{
		{
			Type: migrator.MigrateTypeCreate,
			NewData: objects.Policy{
				Name:  "test-policy-error",
				Table: "error-table",
			},
		},
	}

	createFuncWithError := func(cfg *raiden.Config, param objects.Policy) (objects.Policy, error) {
		return objects.Policy{}, errors.New("policy creation failed")
	}
	
	updateFunc := func(cfg *raiden.Config, param objects.Policy, items objects.UpdatePolicyParam) error {
		return nil
	}
	
	deleteFunc := func(cfg *raiden.Config, param objects.Policy) error {
		return nil
	}
	
	actionFuncs := migrator.MigrateActionFunc[objects.Policy, objects.UpdatePolicyParam]{
		CreateFunc: createFuncWithError,
		UpdateFunc: updateFunc,
		DeleteFunc: deleteFunc,
	}

	errs := migrator.MigratePolicy(cfg, policies, stateChan, actionFuncs)
	assert.NotEmpty(t, errs)
}