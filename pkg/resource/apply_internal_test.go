package resource

import (
	"errors"
	"testing"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/resource/migrator"
	"github.com/sev-2/raiden/pkg/resource/policies"
	"github.com/sev-2/raiden/pkg/resource/roles"
	"github.com/sev-2/raiden/pkg/resource/rpc"
	"github.com/sev-2/raiden/pkg/resource/storages"
	"github.com/sev-2/raiden/pkg/resource/tables"
	"github.com/sev-2/raiden/pkg/resource/types"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunApply_Success(t *testing.T) {
	flags := &Flags{AllowedSchema: "public"}
	config := &raiden.Config{Mode: raiden.BffMode, AllowedTables: "*"}

	appTables := state.ExtractTableResult{Existing: state.ExtractTableItems{{Table: objects.Table{Name: "table1", Schema: "public", Columns: []objects.Column{{Name: "id"}}}, ExtractedPolicies: state.ExtractPolicyResult{Existing: []objects.Policy{{Name: "policy1", Table: "table1", Roles: []string{"role1"}}}}}}}
	appRoles := state.ExtractRoleResult{Existing: []objects.Role{{Name: "role1"}}}
	appRpc := state.ExtractRpcResult{Existing: []objects.Function{{Name: "fn1"}}}
	appStorage := state.ExtractStorageResult{Existing: []state.ExtractStorageItem{{Storage: objects.Bucket{Name: "bucket1"}}}}
	appTypes := state.ExtractTypeResult{Existing: []objects.Type{{Name: "type1"}}}

	resource := &Resource{
		Tables:    []objects.Table{{Name: "table1", Schema: "public", Columns: []objects.Column{{Name: "id"}}}},
		Roles:     []objects.Role{{Name: "role1"}},
		Functions: []objects.Function{{Name: "fn1", Schema: "public"}},
		Storages:  []objects.Bucket{{Name: "bucket1"}},
		Policies:  objects.Policies{{Name: "policy1", Table: "table1", Roles: []string{"role1"}}},
		Types:     []objects.Type{{Name: "type1"}},
	}

	calls := struct {
		migrate   bool
		print     bool
		role      bool
		table     bool
		rpc       bool
		storage   bool
		policy    bool
		typeBuild bool
	}{}

	deps := applyDeps{
		loadNativeRoles: func() (map[string]raiden.Role, error) { return map[string]raiden.Role{}, nil },
		loadState:       func() (*state.State, error) { return &state.State{}, nil },
		extractApp: func(*Flags, *state.State) (state.ExtractTableResult, state.ExtractRoleResult, state.ExtractRpcResult, state.ExtractStorageResult, state.ExtractTypeResult, error) {
			return appTables, appRoles, appRpc, appStorage, appTypes, nil
		},
		loadRemote: func(*Flags, *raiden.Config) (*Resource, error) { return resource, nil },
		migrate: func(*raiden.Config, *state.LocalState, string, *MigrateData) []error {
			calls.migrate = true
			return nil
		},
		buildRoleMigrate: func(state.ExtractRoleResult, []objects.Role) ([]roles.MigrateItem, error) {
			calls.role = true
			return []roles.MigrateItem{{Type: migrator.MigrateTypeCreate}}, nil
		},
		buildTableMigrate: func(state.ExtractTableResult, []objects.Table, []string) ([]tables.MigrateItem, error) {
			calls.table = true
			return []tables.MigrateItem{{Type: migrator.MigrateTypeUpdate}}, nil
		},
		buildRpcMigrate: func(state.ExtractRpcResult, []objects.Function) ([]rpc.MigrateItem, error) {
			calls.rpc = true
			return []rpc.MigrateItem{{Type: migrator.MigrateTypeDelete}}, nil
		},
		buildStorageMigrate: func(state.ExtractStorageResult, []objects.Bucket) ([]storages.MigrateItem, error) {
			calls.storage = true
			return []storages.MigrateItem{{Type: migrator.MigrateTypeIgnore}}, nil
		},
		buildPolicyMigrate: func(state.ExtractPolicyResult, []objects.Policy) ([]policies.MigrateItem, error) {
			calls.policy = true
			return []policies.MigrateItem{}, nil
		},
		buildTypeMigrate: func(state.ExtractTypeResult, []objects.Type) ([]types.MigrateItem, error) {
			calls.typeBuild = true
			return []types.MigrateItem{}, nil
		},
		printReport: func(MigrateData) {
			calls.print = true
		},
	}

	err := runApply(flags, config, deps)
	require.NoError(t, err)
	assert.True(t, calls.migrate)
	assert.True(t, calls.print)
	assert.True(t, calls.role)
	assert.True(t, calls.table)
	assert.True(t, calls.rpc)
	assert.True(t, calls.storage)
	assert.True(t, calls.policy)
	assert.True(t, calls.typeBuild)
}

func TestRunApply_DryRunSkipsMigrate(t *testing.T) {
	flags := &Flags{AllowedSchema: "public", DryRun: true}

	deps := applyDeps{
		loadNativeRoles: func() (map[string]raiden.Role, error) { return map[string]raiden.Role{}, nil },
		loadState:       func() (*state.State, error) { return &state.State{}, nil },
		extractApp: func(*Flags, *state.State) (state.ExtractTableResult, state.ExtractRoleResult, state.ExtractRpcResult, state.ExtractStorageResult, state.ExtractTypeResult, error) {
			return state.ExtractTableResult{}, state.ExtractRoleResult{}, state.ExtractRpcResult{}, state.ExtractStorageResult{}, state.ExtractTypeResult{}, nil
		},
		loadRemote: func(*Flags, *raiden.Config) (*Resource, error) { return &Resource{}, nil },
		migrate: func(*raiden.Config, *state.LocalState, string, *MigrateData) []error {
			t.Fatalf("migrate should not be called in dry run mode")
			return nil
		},
		buildRoleMigrate: func(state.ExtractRoleResult, []objects.Role) ([]roles.MigrateItem, error) { return nil, nil },
		buildTableMigrate: func(state.ExtractTableResult, []objects.Table, []string) ([]tables.MigrateItem, error) {
			return nil, nil
		},
		buildRpcMigrate:     func(state.ExtractRpcResult, []objects.Function) ([]rpc.MigrateItem, error) { return nil, nil },
		buildStorageMigrate: func(state.ExtractStorageResult, []objects.Bucket) ([]storages.MigrateItem, error) { return nil, nil },
		buildPolicyMigrate:  func(state.ExtractPolicyResult, []objects.Policy) ([]policies.MigrateItem, error) { return nil, nil },
		buildTypeMigrate:    func(state.ExtractTypeResult, []objects.Type) ([]types.MigrateItem, error) { return nil, nil },
		printReport:         func(MigrateData) {},
	}

	err := runApply(flags, &raiden.Config{}, deps)
	require.NoError(t, err)
}

func TestRunApply_MigrateError(t *testing.T) {
	deps := applyDeps{
		loadNativeRoles: func() (map[string]raiden.Role, error) { return map[string]raiden.Role{}, nil },
		loadState:       func() (*state.State, error) { return &state.State{}, nil },
		extractApp: func(*Flags, *state.State) (state.ExtractTableResult, state.ExtractRoleResult, state.ExtractRpcResult, state.ExtractStorageResult, state.ExtractTypeResult, error) {
			return state.ExtractTableResult{}, state.ExtractRoleResult{}, state.ExtractRpcResult{}, state.ExtractStorageResult{}, state.ExtractTypeResult{}, nil
		},
		loadRemote:       func(*Flags, *raiden.Config) (*Resource, error) { return &Resource{}, nil },
		buildRoleMigrate: func(state.ExtractRoleResult, []objects.Role) ([]roles.MigrateItem, error) { return nil, nil },
		buildTableMigrate: func(state.ExtractTableResult, []objects.Table, []string) ([]tables.MigrateItem, error) {
			return nil, nil
		},
		buildRpcMigrate:     func(state.ExtractRpcResult, []objects.Function) ([]rpc.MigrateItem, error) { return nil, nil },
		buildStorageMigrate: func(state.ExtractStorageResult, []objects.Bucket) ([]storages.MigrateItem, error) { return nil, nil },
		buildPolicyMigrate:  func(state.ExtractPolicyResult, []objects.Policy) ([]policies.MigrateItem, error) { return nil, nil },
		buildTypeMigrate:    func(state.ExtractTypeResult, []objects.Type) ([]types.MigrateItem, error) { return nil, nil },
		migrate: func(*raiden.Config, *state.LocalState, string, *MigrateData) []error {
			return []error{errors.New("boom")}
		},
		printReport: func(MigrateData) {},
	}

	err := runApply(&Flags{}, &raiden.Config{}, deps)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "boom")
}

func TestRunApply_LoadStateMissingReturnsError(t *testing.T) {
	deps := applyDeps{
		loadNativeRoles: func() (map[string]raiden.Role, error) { return map[string]raiden.Role{}, nil },
		loadState:       func() (*state.State, error) { return nil, nil },
		printReport:     func(MigrateData) {},
		migrate:         func(*raiden.Config, *state.LocalState, string, *MigrateData) []error { return nil },
		loadRemote:      func(*Flags, *raiden.Config) (*Resource, error) { return &Resource{}, nil },
		extractApp: func(*Flags, *state.State) (state.ExtractTableResult, state.ExtractRoleResult, state.ExtractRpcResult, state.ExtractStorageResult, state.ExtractTypeResult, error) {
			return state.ExtractTableResult{}, state.ExtractRoleResult{}, state.ExtractRpcResult{}, state.ExtractStorageResult{}, state.ExtractTypeResult{}, nil
		},
		buildRoleMigrate: func(state.ExtractRoleResult, []objects.Role) ([]roles.MigrateItem, error) { return nil, nil },
		buildTableMigrate: func(state.ExtractTableResult, []objects.Table, []string) ([]tables.MigrateItem, error) {
			return nil, nil
		},
		buildRpcMigrate:     func(state.ExtractRpcResult, []objects.Function) ([]rpc.MigrateItem, error) { return nil, nil },
		buildStorageMigrate: func(state.ExtractStorageResult, []objects.Bucket) ([]storages.MigrateItem, error) { return nil, nil },
		buildPolicyMigrate:  func(state.ExtractPolicyResult, []objects.Policy) ([]policies.MigrateItem, error) { return nil, nil },
		buildTypeMigrate:    func(state.ExtractTypeResult, []objects.Type) ([]types.MigrateItem, error) { return nil, nil },
	}

	err := runApply(&Flags{}, &raiden.Config{}, deps)
	require.EqualError(t, err, "state file is not found, please run raiden imports first")
}

func TestRunApply_LoadNativeRoleError(t *testing.T) {
	dummyErr := errors.New("native error")
	deps := applyDeps{
		loadNativeRoles: func() (map[string]raiden.Role, error) { return nil, dummyErr },
		loadState:       func() (*state.State, error) { return &state.State{}, nil },
		extractApp: func(*Flags, *state.State) (state.ExtractTableResult, state.ExtractRoleResult, state.ExtractRpcResult, state.ExtractStorageResult, state.ExtractTypeResult, error) {
			return state.ExtractTableResult{}, state.ExtractRoleResult{}, state.ExtractRpcResult{}, state.ExtractStorageResult{}, state.ExtractTypeResult{}, nil
		},
		loadRemote:       func(*Flags, *raiden.Config) (*Resource, error) { return &Resource{}, nil },
		migrate:          func(*raiden.Config, *state.LocalState, string, *MigrateData) []error { return nil },
		printReport:      func(MigrateData) {},
		buildRoleMigrate: func(state.ExtractRoleResult, []objects.Role) ([]roles.MigrateItem, error) { return nil, nil },
		buildTableMigrate: func(state.ExtractTableResult, []objects.Table, []string) ([]tables.MigrateItem, error) {
			return nil, nil
		},
		buildRpcMigrate:     func(state.ExtractRpcResult, []objects.Function) ([]rpc.MigrateItem, error) { return nil, nil },
		buildStorageMigrate: func(state.ExtractStorageResult, []objects.Bucket) ([]storages.MigrateItem, error) { return nil, nil },
		buildPolicyMigrate:  func(state.ExtractPolicyResult, []objects.Policy) ([]policies.MigrateItem, error) { return nil, nil },
		buildTypeMigrate:    func(state.ExtractTypeResult, []objects.Type) ([]types.MigrateItem, error) { return nil, nil },
	}

	err := runApply(&Flags{}, &raiden.Config{}, deps)
	assert.Equal(t, dummyErr, err)
}

func TestRunApply_BuildTableMigrateError(t *testing.T) {
	expectedErr := errors.New("table migrate build failed")

	deps := applyDeps{
		loadNativeRoles: func() (map[string]raiden.Role, error) { return map[string]raiden.Role{}, nil },
		loadState:       func() (*state.State, error) { return &state.State{}, nil },
		extractApp: func(*Flags, *state.State) (state.ExtractTableResult, state.ExtractRoleResult, state.ExtractRpcResult, state.ExtractStorageResult, state.ExtractTypeResult, error) {
			return state.ExtractTableResult{}, state.ExtractRoleResult{}, state.ExtractRpcResult{}, state.ExtractStorageResult{}, state.ExtractTypeResult{}, nil
		},
		loadRemote:       func(*Flags, *raiden.Config) (*Resource, error) { return &Resource{}, nil },
		buildRoleMigrate: func(state.ExtractRoleResult, []objects.Role) ([]roles.MigrateItem, error) { return nil, nil },
		buildTableMigrate: func(state.ExtractTableResult, []objects.Table, []string) ([]tables.MigrateItem, error) {
			return nil, expectedErr
		},
		buildRpcMigrate:     func(state.ExtractRpcResult, []objects.Function) ([]rpc.MigrateItem, error) { return nil, nil },
		buildStorageMigrate: func(state.ExtractStorageResult, []objects.Bucket) ([]storages.MigrateItem, error) { return nil, nil },
		buildPolicyMigrate:  func(state.ExtractPolicyResult, []objects.Policy) ([]policies.MigrateItem, error) { return nil, nil },
		buildTypeMigrate:    func(state.ExtractTypeResult, []objects.Type) ([]types.MigrateItem, error) { return nil, nil },
		migrate:             func(*raiden.Config, *state.LocalState, string, *MigrateData) []error { return nil },
		printReport:         func(MigrateData) {},
	}

	err := runApply(&Flags{AllowedSchema: "public"}, &raiden.Config{AllowedTables: "*", Mode: raiden.BffMode}, deps)
	require.Equal(t, expectedErr, err)
}
