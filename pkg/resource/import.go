package resource

import (
	"errors"
	"strings"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/logger"
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
	logger.Info("import : load native role")
	mapNativeRole, err := loadMapNativeRole()
	if err != nil {
		return err
	}

	// load supabase resource
	logger.Info("import : load table, role, function, model and policy from supabase")
	spResource, err := Load(flags, config)
	if err != nil {
		return err
	}

	// create import state
	nativeStateRoles := filterIsNativeRole(mapNativeRole, spResource.Roles)

	// filter table for with allowed schema
	spResource.Tables = filterTableBySchema(spResource.Tables, strings.Split(flags.AllowedSchema, ",")...)
	spResource.Functions = filterFunctionBySchema(spResource.Functions, strings.Split(flags.AllowedSchema, ",")...)
	spResource.Roles = filterUserRole(spResource.Roles, mapNativeRole)

	// load app resource
	logger.Info("import : load local state")
	latestState, err := loadState()
	if err != nil {
		return err
	}

	logger.Info("import : extract load table, role, function, model and policy from local state")
	appTables, appRoles, appRpcFunctions, err := extractAppResource(flags, latestState)
	if err != nil {
		return err
	}

	importState := ResourceState{
		State: state.State{
			Roles: nativeStateRoles,
		},
	}

	// compare resource
	if (flags.All() || flags.ModelsOnly) && len(appTables.Existing) > 0 {
		logger.Info("import : compare table")
		// compare table
		var compareTables []objects.Table
		for i := range appTables.Existing {
			et := appTables.Existing[i]
			compareTables = append(compareTables, et.Table)
		}

		if err := runImportCompareTable(spResource.Tables, compareTables); err != nil {
			return err
		}
	}

	if (flags.All() || flags.RolesOnly) && len(appRoles.Existing) > 0 {
		logger.Info("import : compare roles")
		if err := runImportCompareRoles(spResource.Roles, appRoles.Existing); err != nil {
			return err
		}
	}

	if (flags.All() || flags.RpcOnly) && len(appRpcFunctions.Existing) > 0 {
		logger.Info("import : compare rpc")
		if err := runImportCompareRpcFunctions(spResource.Functions, appRpcFunctions.Existing); err != nil {
			return err
		}
	}

	// generate resource
	if err := generateResource(config, &importState, flags.ProjectPath, spResource); err != nil {
		return err
	}

	logger.Infof(`imports result - table : %v roles : %v policy : %v function : %v`, len(spResource.Tables), len(spResource.Roles), len(spResource.Policies), len(spResource.Functions))
	return nil
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
	diffResult, err := CompareRoles(supabaseRoles, appRoles, CompareModeImport)
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
