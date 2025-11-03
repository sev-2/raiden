package resource

import (
	"errors"
	"testing"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunImport_GenerateSuccess(t *testing.T) {
	flags := &Flags{AllowedSchema: "public", ProjectPath: "proj"}
	config := &raiden.Config{Mode: raiden.BffMode, AllowedTables: "*", ProjectName: "proj"}

	remoteResource := &Resource{
		Tables:    []objects.Table{{ID: 1, Name: "table1", Schema: "public"}},
		Roles:     []objects.Role{{Name: "role_app"}},
		Functions: []objects.Function{{Name: "func1", Schema: "public"}},
		Storages:  []objects.Bucket{{Name: "bucket1"}},
		Types:     []objects.Type{{Name: "type1"}},
	}

	appTables := state.ExtractTableResult{
		Existing: state.ExtractTableItems{{
			Table:          objects.Table{ID: 1, Name: "table1", Schema: "public"},
			ValidationTags: state.ModelValidationTag{"existing": "true"},
		}},
		New: state.ExtractTableItems{{
			Table:          objects.Table{Name: "table2", Schema: "public"},
			ValidationTags: state.ModelValidationTag{"new": "true"},
		}},
	}
	appRoles := state.ExtractRoleResult{Existing: []objects.Role{{Name: "role_app"}}}
	appRpc := state.ExtractRpcResult{Existing: []objects.Function{{Name: "func1"}}}
	appStorage := state.ExtractStorageResult{Existing: []state.ExtractStorageItem{{Storage: objects.Bucket{Name: "bucket1"}}}}
	appTypes := state.ExtractTypeResult{Existing: []objects.Type{{Name: "type1"}}}

	called := struct {
		compareTypes    bool
		compareTables   bool
		compareRoles    bool
		compareRpc      bool
		compareStorages bool
		generate        bool
		print           bool
	}{}

	var capturedReport ImportReport
	deps := importDeps{
		loadNativeRoles: func() (map[string]raiden.Role, error) {
			return map[string]raiden.Role{}, nil
		},
		loadRemote: func(*Flags, *raiden.Config) (*Resource, error) {
			return remoteResource, nil
		},
		loadState: func() (*state.State, error) {
			return &state.State{}, nil
		},
		extractApp: func(*Flags, *state.State) (state.ExtractTableResult, state.ExtractRoleResult, state.ExtractRpcResult, state.ExtractStorageResult, state.ExtractTypeResult, error) {
			return appTables, appRoles, appRpc, appStorage, appTypes, nil
		},
		compareTypes: func(remote []objects.Type, existing []objects.Type) error {
			called.compareTypes = true
			require.Equal(t, remoteResource.Types, remote)
			require.Equal(t, appTypes.Existing, existing)
			return nil
		},
		compareTables: func(remote []objects.Table, existing []objects.Table) error {
			called.compareTables = true
			require.Equal(t, remoteResource.Tables, remote)
			require.Len(t, existing, 1)
			return nil
		},
		compareRoles: func(remote []objects.Role, existing []objects.Role) error {
			called.compareRoles = true
			require.Equal(t, remoteResource.Roles, remote)
			require.Equal(t, appRoles.Existing, existing)
			return nil
		},
		compareRpc: func(remote []objects.Function, existing []objects.Function) error {
			called.compareRpc = true
			require.Equal(t, remoteResource.Functions, remote)
			require.Equal(t, appRpc.Existing, existing)
			return nil
		},
		compareStorages: func(remote []objects.Bucket, existing []objects.Bucket) error {
			called.compareStorages = true
			require.Equal(t, remoteResource.Storages, remote)
			require.Len(t, existing, 1)
			return nil
		},
		updateStateOnly: func(*state.LocalState, *Resource, map[string]state.ModelValidationTag) error {
			t.Fatalf("unexpected updateStateOnly call")
			return nil
		},
		generate: func(cfg *raiden.Config, ls *state.LocalState, projectPath string, res *Resource, tags map[string]state.ModelValidationTag, generateController bool) error {
			called.generate = true
			require.Equal(t, config, cfg)
			require.Equal(t, "proj", projectPath)
			require.Equal(t, remoteResource, res)
			require.True(t, generateController)
			require.Equal(t, 2, len(tags))
			return nil
		},
		printReport: func(report ImportReport, dryRun bool) {
			called.print = true
			require.False(t, dryRun)
			capturedReport = report
		},
	}
	flags.GenerateController = true
	flags.ModelsOnly = false

	err := runImport(flags, config, deps)
	require.NoError(t, err)
	assert.True(t, called.compareTypes)
	assert.True(t, called.compareTables)
	assert.True(t, called.compareRoles)
	assert.True(t, called.compareRpc)
	assert.True(t, called.compareStorages)
	assert.True(t, called.generate)
	assert.True(t, called.print)
	assert.GreaterOrEqual(t, capturedReport.Table, 0)
}

func TestRunImport_DryRunWithError(t *testing.T) {
	flags := &Flags{AllowedSchema: "public", DryRun: true, ModelsOnly: true}
	config := &raiden.Config{Mode: raiden.BffMode, AllowedTables: "*"}

	appTables := state.ExtractTableResult{
		Existing: state.ExtractTableItems{{Table: objects.Table{Name: "table1", Schema: "public"}}},
	}

	deps := importDeps{
		loadNativeRoles: func() (map[string]raiden.Role, error) { return map[string]raiden.Role{}, nil },
		loadRemote: func(*Flags, *raiden.Config) (*Resource, error) {
			return &Resource{Tables: []objects.Table{{Name: "table1", Schema: "public"}}}, nil
		},
		loadState: func() (*state.State, error) { return &state.State{}, nil },
		extractApp: func(*Flags, *state.State) (state.ExtractTableResult, state.ExtractRoleResult, state.ExtractRpcResult, state.ExtractStorageResult, state.ExtractTypeResult, error) {
			return appTables, state.ExtractRoleResult{}, state.ExtractRpcResult{}, state.ExtractStorageResult{}, state.ExtractTypeResult{}, nil
		},
		compareTypes: func([]objects.Type, []objects.Type) error { return nil },
		compareTables: func([]objects.Table, []objects.Table) error {
			return errors.New("compare tables failed")
		},
		compareRoles:    func([]objects.Role, []objects.Role) error { return nil },
		compareRpc:      func([]objects.Function, []objects.Function) error { return nil },
		compareStorages: func([]objects.Bucket, []objects.Bucket) error { return nil },
		updateStateOnly: func(*state.LocalState, *Resource, map[string]state.ModelValidationTag) error { return nil },
		generate: func(*raiden.Config, *state.LocalState, string, *Resource, map[string]state.ModelValidationTag, bool) error {
			t.Fatalf("unexpected generate call")
			return nil
		},
		printReport: func(ImportReport, bool) { t.Fatalf("report should not be printed when there are dry run errors") },
	}

	err := runImport(flags, config, deps)
	require.NoError(t, err)
}

func TestRunImport_ForceImportSkipsComparisons(t *testing.T) {
	flags := &Flags{AllowedSchema: "public", ForceImport: true}
	config := &raiden.Config{Mode: raiden.BffMode, AllowedTables: "*"}

	called := struct {
		compare bool
	}{}

	deps := importDeps{
		loadNativeRoles: func() (map[string]raiden.Role, error) { return map[string]raiden.Role{}, nil },
		loadRemote:      func(*Flags, *raiden.Config) (*Resource, error) { return &Resource{}, nil },
		loadState:       func() (*state.State, error) { return &state.State{}, nil },
		extractApp: func(*Flags, *state.State) (state.ExtractTableResult, state.ExtractRoleResult, state.ExtractRpcResult, state.ExtractStorageResult, state.ExtractTypeResult, error) {
			return state.ExtractTableResult{}, state.ExtractRoleResult{}, state.ExtractRpcResult{}, state.ExtractStorageResult{}, state.ExtractTypeResult{}, nil
		},
		compareTypes: func([]objects.Type, []objects.Type) error {
			called.compare = true
			return nil
		},
		compareTables: func([]objects.Table, []objects.Table) error {
			called.compare = true
			return nil
		},
		compareRoles: func([]objects.Role, []objects.Role) error {
			called.compare = true
			return nil
		},
		compareRpc: func([]objects.Function, []objects.Function) error {
			called.compare = true
			return nil
		},
		compareStorages: func([]objects.Bucket, []objects.Bucket) error {
			called.compare = true
			return nil
		},
		updateStateOnly: func(*state.LocalState, *Resource, map[string]state.ModelValidationTag) error { return nil },
		generate: func(*raiden.Config, *state.LocalState, string, *Resource, map[string]state.ModelValidationTag, bool) error {
			return nil
		},
		printReport: func(ImportReport, bool) {},
	}

	err := runImport(flags, config, deps)
	require.NoError(t, err)
	assert.False(t, called.compare, "comparisons should be skipped when force import is enabled")
}

func TestRunImport_UpdateStateOnly(t *testing.T) {
	flags := &Flags{AllowedSchema: "public", UpdateStateOnly: true}
	config := &raiden.Config{Mode: raiden.BffMode, AllowedTables: "*"}

	updateCalled := false
	deps := importDeps{
		loadNativeRoles: func() (map[string]raiden.Role, error) { return map[string]raiden.Role{}, nil },
		loadRemote:      func(*Flags, *raiden.Config) (*Resource, error) { return &Resource{}, nil },
		loadState:       func() (*state.State, error) { return &state.State{}, nil },
		extractApp: func(*Flags, *state.State) (state.ExtractTableResult, state.ExtractRoleResult, state.ExtractRpcResult, state.ExtractStorageResult, state.ExtractTypeResult, error) {
			return state.ExtractTableResult{}, state.ExtractRoleResult{}, state.ExtractRpcResult{}, state.ExtractStorageResult{}, state.ExtractTypeResult{}, nil
		},
		compareTypes:    func([]objects.Type, []objects.Type) error { return nil },
		compareTables:   func([]objects.Table, []objects.Table) error { return nil },
		compareRoles:    func([]objects.Role, []objects.Role) error { return nil },
		compareRpc:      func([]objects.Function, []objects.Function) error { return nil },
		compareStorages: func([]objects.Bucket, []objects.Bucket) error { return nil },
		updateStateOnly: func(*state.LocalState, *Resource, map[string]state.ModelValidationTag) error {
			updateCalled = true
			return nil
		},
		generate: func(*raiden.Config, *state.LocalState, string, *Resource, map[string]state.ModelValidationTag, bool) error {
			t.Fatalf("generate should not be called when updateStateOnly is set")
			return nil
		},
		printReport: func(ImportReport, bool) {},
	}

	err := runImport(flags, config, deps)
	require.NoError(t, err)
	assert.True(t, updateCalled)
}

func TestRunImport_LoadNativeRoleError(t *testing.T) {
	dummyErr := errors.New("boom")
	deps := importDeps{
		loadNativeRoles: func() (map[string]raiden.Role, error) { return nil, dummyErr },
		loadRemote:      func(*Flags, *raiden.Config) (*Resource, error) { t.Fatalf("should not be called"); return nil, nil },
		loadState:       func() (*state.State, error) { return &state.State{}, nil },
		extractApp: func(*Flags, *state.State) (state.ExtractTableResult, state.ExtractRoleResult, state.ExtractRpcResult, state.ExtractStorageResult, state.ExtractTypeResult, error) {
			return state.ExtractTableResult{}, state.ExtractRoleResult{}, state.ExtractRpcResult{}, state.ExtractStorageResult{}, state.ExtractTypeResult{}, nil
		},
		compareTypes:    func([]objects.Type, []objects.Type) error { return nil },
		compareTables:   func([]objects.Table, []objects.Table) error { return nil },
		compareRoles:    func([]objects.Role, []objects.Role) error { return nil },
		compareRpc:      func([]objects.Function, []objects.Function) error { return nil },
		compareStorages: func([]objects.Bucket, []objects.Bucket) error { return nil },
		updateStateOnly: func(*state.LocalState, *Resource, map[string]state.ModelValidationTag) error { return nil },
		generate: func(*raiden.Config, *state.LocalState, string, *Resource, map[string]state.ModelValidationTag, bool) error {
			return nil
		},
		printReport: func(ImportReport, bool) {},
	}

	err := runImport(&Flags{}, &raiden.Config{}, deps)
	assert.Equal(t, dummyErr, err)
}

func TestRunImport_CompareTablesError(t *testing.T) {
	flags := &Flags{AllowedSchema: "public"}
	expectedErr := errors.New("diff failed")

	deps := importDeps{
		loadNativeRoles: func() (map[string]raiden.Role, error) { return map[string]raiden.Role{}, nil },
		loadRemote: func(*Flags, *raiden.Config) (*Resource, error) {
			return &Resource{Tables: []objects.Table{{Name: "table1", Schema: "public"}}}, nil
		},
		loadState: func() (*state.State, error) { return &state.State{}, nil },
		extractApp: func(*Flags, *state.State) (state.ExtractTableResult, state.ExtractRoleResult, state.ExtractRpcResult, state.ExtractStorageResult, state.ExtractTypeResult, error) {
			return state.ExtractTableResult{
				Existing: state.ExtractTableItems{{
					Table: objects.Table{Name: "table1", Schema: "public"},
				}},
			}, state.ExtractRoleResult{}, state.ExtractRpcResult{}, state.ExtractStorageResult{}, state.ExtractTypeResult{}, nil
		},
		compareTypes:    func([]objects.Type, []objects.Type) error { return nil },
		compareTables:   func([]objects.Table, []objects.Table) error { return expectedErr },
		compareRoles:    func([]objects.Role, []objects.Role) error { return nil },
		compareRpc:      func([]objects.Function, []objects.Function) error { return nil },
		compareStorages: func([]objects.Bucket, []objects.Bucket) error { return nil },
		updateStateOnly: func(*state.LocalState, *Resource, map[string]state.ModelValidationTag) error {
			t.Fatalf("should not update state")
			return nil
		},
		generate: func(*raiden.Config, *state.LocalState, string, *Resource, map[string]state.ModelValidationTag, bool) error {
			t.Fatalf("should not generate")
			return nil
		},
		printReport: func(ImportReport, bool) { t.Fatalf("should not print report") },
	}

	err := runImport(flags, &raiden.Config{}, deps)
	require.Equal(t, expectedErr, err)
}
