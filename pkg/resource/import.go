package resource

import (
	"errors"
	"strings"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase/objects"
)

// List of import resource
// [x] import table, relation, column specification and acl
// [ ] delete unused models
// [x] import role
// [ ] delete unused role
// [x] import function
// [ ] delete  unused function
func Import(flags *Flags, config *raiden.Config) error {
	// load map native role
	mapNativeRole, err := loadMapNativeRole()
	if err != nil {
		return err
	}

	// load supabase resource
	resource, err := Load(flags, config)
	if err != nil {
		return err
	}

	// create import state
	nativeStateRoles := filterIsNativeRole(mapNativeRole, resource.Roles)
	if err != nil {
		return err
	}

	// filter table for with allowed schema
	resource.Tables = filterTableBySchema(resource.Tables, strings.Split(flags.AllowedSchema, ",")...)
	resource.Functions = filterFunctionBySchema(resource.Functions, strings.Split(flags.AllowedSchema, ",")...)
	resource.Roles = filterUserRole(resource.Roles, mapNativeRole)

	// load app resource
	latestState, err := loadAppResource()
	if err != nil {
		return err
	}
	appTables, appRoles, appRpcFunctions, err := extractAppResourceState(flags, latestState)
	if err != nil {
		return err
	}

	// compare
	if (flags.All() || flags.ModelsOnly) && len(appTables.ExistingTable) > 0 {
		if err := runImportCompareTable(resource.Tables, appTables.ExistingTable); err != nil {
			return err
		}
	}

	if (flags.All() || flags.RolesOnly) && len(appRoles) > 0 {
		if err := runImportCompareRoles(resource.Roles, appRoles); err != nil {
			return err
		}
	}

	if (flags.All() || flags.RpcOnly) && len(appRpcFunctions) > 0 {
		if err := runImportCompareRpcFunctions(resource.Functions, appRpcFunctions); err != nil {
			return err
		}
	}

	importState := resourceState{
		State: state.State{
			Roles: nativeStateRoles,
		},
	}

	// generate resource
	return generateResource(config, &importState, flags.ProjectPath, resource)
}

func runImportCompareTable(supabaseTable []objects.Table, appTable []objects.Table) error {
	diffResult, err := CompareTables(supabaseTable, appTable, CompareModeImport)
	if err != nil {
		return err
	}

	if len(diffResult) > 0 {
		for i := range diffResult {
			d := diffResult[i]
			PrintDiff("table", d.SourceResource, d.TargetResource, d.Name)
		}
		return errors.New("import tables is canceled, you have conflict table. please fix it first")
	}

	return nil
}

func runImportCompareRoles(supabaseRoles []objects.Role, appRoles []objects.Role) error {
	diffResult, err := CompareRoles(supabaseRoles, appRoles)
	if err != nil {
		return err
	}

	if len(diffResult) > 0 {
		for i := range diffResult {
			d := diffResult[i]
			PrintDiff("role", d.SourceResource, d.TargetResource, d.Name)
		}
		return errors.New("import roles is canceled, you have conflict role. please fix it first")
	}

	return nil
}

func runImportCompareRpcFunctions(supabaseFn []objects.Function, appFn []objects.Function) error {
	diffResult, err := CompareRpcFunctions(supabaseFn, appFn)
	if err != nil {
		return err
	}

	if len(diffResult) > 0 {
		for i := range diffResult {
			d := diffResult[i]
			PrintDiff("rpc function", d.SourceResource, d.TargetResource, d.Name)
		}
		return errors.New("import rpc function is canceled, you have conflict rpc function. please fix it first")
	}

	return nil
}
