package resource

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/go-hclog"
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase/objects"
)

var ImportLogger hclog.Logger = logger.HcLog().Named("import")

// List of import resource
// [x] import table, relation, column specification and acl
// [x] import role
// [x] import function
// [x] import storage
func Import(flags *Flags, config *raiden.Config) error {

	// load map native role
	ImportLogger.Info("load Native log")
	mapNativeRole, err := loadMapNativeRole()
	if err != nil {
		return err
	}

	// load supabase resource
	ImportLogger.Info("start - load resource from supabase")
	spResource, err := Load(flags, config)
	if err != nil {
		return err
	}
	ImportLogger.Info("finish - load resource from supabase")

	// create import state
	ImportLogger.Debug("get native roles")
	nativeStateRoles := filterIsNativeRole(mapNativeRole, spResource.Roles)

	// filter table for with allowed schema
	ImportLogger.Debug("start - filter table and function by allowed schema", "allowed-schema", flags.AllowedSchema)
	ImportLogger.Trace("filter table by schema")
	spResource.Tables = filterTableBySchema(spResource.Tables, strings.Split(flags.AllowedSchema, ",")...)

	ImportLogger.Trace("filter function by schema")
	spResource.Functions = filterFunctionBySchema(spResource.Functions, strings.Split(flags.AllowedSchema, ",")...)
	ImportLogger.Debug("finish - filter table and function by allowed schema")

	ImportLogger.Trace("remove native role for supabase list role")
	spResource.Roles = filterUserRole(spResource.Roles, mapNativeRole)

	// load app resource
	ImportLogger.Info("start - load resource from local state")
	latestState, err := loadState()
	if err != nil {
		return err
	}
	ImportLogger.Info("finish - load resource from local state")

	ImportLogger.Info("start - extract data from local state")
	appTables, appRoles, appRpcFunctions, appStorage, err := extractAppResource(flags, latestState)
	if err != nil {
		return err
	}
	ImportLogger.Info("finish - extract data from local state")

	importState := ResourceState{
		State: state.State{
			Roles: nativeStateRoles,
		},
	}

	// compare resource
	ImportLogger.Info("start - compare supabase resource and local resource")
	if (flags.All() || flags.ModelsOnly) && len(appTables.Existing) > 0 {
		ImportLogger.Debug("start - compare table")
		// compare table
		var compareTables []objects.Table
		for i := range appTables.Existing {
			et := appTables.Existing[i]
			compareTables = append(compareTables, et.Table)
		}

		if err := runImportCompareTable(spResource.Tables, compareTables); err != nil {
			return err
		}
		ImportLogger.Debug("finish - compare table")
	}

	if (flags.All() || flags.RolesOnly) && len(appRoles.Existing) > 0 {
		ImportLogger.Debug("start - compare role")
		if err := runImportCompareRoles(spResource.Roles, appRoles.Existing); err != nil {
			return err
		}
		ImportLogger.Debug("finish - compare role")
	}

	if (flags.All() || flags.RpcOnly) && len(appRpcFunctions.Existing) > 0 {
		ImportLogger.Debug("start - compare rpc")
		if err := runImportCompareRpcFunctions(spResource.Functions, appRpcFunctions.Existing); err != nil {
			return err
		}
		ImportLogger.Debug("finish - compare rpc")
	}

	if (flags.All() || flags.StoragesOnly) && len(appStorage.Existing) > 0 {
		ImportLogger.Debug("start - compare storage")
		if err := runImportCompareStorage(spResource.Storages, appStorage.Existing); err != nil {
			return err
		}
		ImportLogger.Debug("finish - compare storage")
	}
	ImportLogger.Info("finish - compare supabase resource and local resource")

	// generate resource
	if err := generateResource(config, &importState, flags.ProjectPath, spResource); err != nil {
		return err
	}

	// logger.Infof(`imports result - table : %v roles : %v policy : %v function : %v`, len(spResource.Tables), len(spResource.Roles), len(spResource.Policies), len(spResource.Functions))
	return nil
}

func runImportCompareTable(supabaseTable []objects.Table, appTable []objects.Table) error {
	diffResult, err := CompareTables(supabaseTable, appTable)
	if err != nil {
		return err
	}

	if len(diffResult) > 0 {
		for i := range diffResult {
			d := diffResult[i]
			PrintDiff("table", d.SourceResource, d.TargetResource, d.Name)
		}
		return errors.New("canceled import process,you have conflict table. please fix it first")
	}

	return nil
}

func runImportCompareRoles(supabaseRoles []objects.Role, appRoles []objects.Role) error {
	diffResult, err := CompareRoles(supabaseRoles, appRoles)
	if err != nil {
		return err
	}
	return printDiffResult(diffResult, PrintDiffTypeRole)
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
		return errors.New("canceled import process, you have conflict rpc function. please fix it first")
	}

	return nil
}

func runImportCompareStorage(supabaseStorage []objects.Bucket, appStorages []objects.Bucket) error {
	diffResult, err := CompareStorage(supabaseStorage, appStorages)
	if err != nil {
		return err
	}
	return printDiffResult(diffResult, PrintDiffTypeStorage)
}

func printDiffResult[T any, P any](diffResult []CompareDiffResult[T, P], printDiff PrintDiffType) error {
	if len(diffResult) == 0 {
		return nil
	}

	isConflict := false
	for i := range diffResult {
		d := diffResult[i]
		if d.IsConflict {
			PrintDiffResource(printDiff, d)
			if !isConflict {
				isConflict = true
			}
		}
	}

	if isConflict {
		return fmt.Errorf("canceled import process, you have conflict in %s. please fix it first", printDiff)
	}

	return nil
}
