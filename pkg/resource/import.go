package resource

import (
	"errors"
	"strings"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/cli"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase/objects"
)

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
	appTables, appRoles, appRpcFunctions, err := loadAppResource(flags)
	if err != nil {
		return err
	}

	// compare
	if (flags.LoadAll() || flags.ModelsOnly) && len(appTables) > 0 {
		if err := runImportCompareTable(resource.Tables, appTables); err != nil {
			return err
		}
	}

	if (flags.LoadAll() || flags.RolesOnly) && len(appRoles) > 0 {
		if err := runImportCompareRoles(resource.Roles, appRoles); err != nil {
			return err
		}
	}

	if (flags.LoadAll() || flags.RpcOnly) && len(appRpcFunctions) > 0 {
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
	diffResult, err := CompareTables(supabaseTable, appTable)
	if err != nil {
		return err
	}

	if len(diffResult) > 0 {
		for i := range diffResult {
			d := diffResult[i]
			cli.PrintDiff("table", d.SourceResource, d.TargetResource, d.Name)
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
			cli.PrintDiff("role", d.SourceResource, d.TargetResource, d.Name)
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
			cli.PrintDiff("rpc function", d.SourceResource, d.TargetResource, d.Name)
		}
		return errors.New("import rpc function is canceled, you have conflict rpc function. please fix it first")
	}

	return nil
}
