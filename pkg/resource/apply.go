package resource

import (
	"errors"
	"strings"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/supabase/objects"
)

// Migrate resource :
//
// [ ] migrate table
//
//	[x] create table name, schema and column
//	[x] create table rls enable
//	[x] create table rls force
//	[x] create table with primary key
//	[ ] create table with relation
//	[ ] create table with acl (rls)
//	[x] delete table
//	[x] update table name, schema
//	[x] update table rls enable
//	[x] update table rls force
//	[x] update table column add new column
//	[x] update table column delete column
//	[x] update table column set default schema
//	[x] update table column set set data type
//	[x] update table column set unique column
//	[x] update table column set nullable
//
// [ ] migrate role
//
//	[ ] create new role
//	[ ] delete role
//	[ ] update name
//	[ ] update connection limit
//	[ ] update inherit role
//	[ ] update is replication role
//	[ ] update is super user
//	[ ] update can bypass rls
//	[ ] update can create db
//	[ ] update can create role
//	[ ] update can login
//
// [ ] migrate function
func Apply(flags *Flags, config *raiden.Config) error {
	// declare default variable
	var migrateResource MigrateData
	var importState resourceState

	// load map native role
	logger.Debug("Load native role")
	mapNativeRole, err := loadMapNativeRole()
	if err != nil {
		return err
	}

	// load supabase resource
	logger.Debug(strings.Repeat("=", 5), " Load supabase resource ")
	resource, err := Load(flags, config)
	if err != nil {
		return err
	}
	logger.Debug(strings.Repeat("=", 5))

	// filter table for with allowed schema
	logger.Debug("Filter supabase resource by allowed schema and make filter user defined role")
	resource.Tables = filterTableBySchema(resource.Tables, strings.Split(flags.AllowedSchema, ",")...)
	resource.Functions = filterFunctionBySchema(resource.Functions, strings.Split(flags.AllowedSchema, ",")...)
	resource.Roles = filterUserRole(resource.Roles, mapNativeRole)

	// load app resource
	logger.Debug(strings.Repeat("=", 5), " Load app resource from state")
	latestState, err := loadAppResource()
	if err != nil {
		return err
	}

	if latestState == nil {
		return errors.New("state file is not found, please run raiden import first")
	}

	importState.State = *latestState

	appTables, appRoles, appRpcFunctions, err := extractAppResourceState(flags, latestState)
	if err != nil {
		return err
	}
	logger.Debug(strings.Repeat("=", 5))

	if flags.All() || flags.ModelsOnly {
		if len(appTables.NewTable) > 0 {
			for i := range appTables.NewTable {
				t := appTables.NewTable[i]
				migrateResource.Tables = append(migrateResource.Tables, MigrateItem[objects.Table, objects.UpdateTableItem]{
					Type:    MigrateTypeCreate,
					NewData: t,
				})
			}
		}

		if rs, err := runApplyCompareTable(resource.Tables, appTables.ExistingTable); err != nil {
			return err
		} else {
			migrateResource.Tables = append(migrateResource.Tables, rs...)
		}

		if len(appTables.DeleteTable) > 0 {
			for i := range appTables.DeleteTable {
				t := appTables.DeleteTable[i]
				isExist := false
				for i := range resource.Tables {
					tt := resource.Tables[i]
					if tt.Name == t.Name {
						isExist = true
						break
					}
				}

				if isExist {
					migrateResource.Tables = append(migrateResource.Tables, MigrateItem[objects.Table, objects.UpdateTableItem]{
						Type:    MigrateTypeDelete,
						OldData: t,
					})
				}
			}
		}
	}

	logger.Debug(strings.Repeat("=", 5), " Table to migrate")
	for i := range migrateResource.Tables {
		t := migrateResource.Tables[i]
		logger.Debugf("- MigrateType : %s, NewTable : %s, OldTable : %s , update items : %s", t.Type, t.NewData.Name, t.OldData.Name, t.MigrationItems)
	}
	logger.Debug(strings.Repeat("=", 5))

	if (flags.All() || flags.RolesOnly) && len(appRoles) > 0 {
		if err := runApplyCompareRoles(resource.Roles, appRoles); err != nil {
			return err
		}
	}

	if (flags.All() || flags.RpcOnly) && len(appRpcFunctions) > 0 {
		if err := runApplyCompareRpcFunctions(resource.Functions, appRpcFunctions); err != nil {
			return err
		}
	}

	migrateErr := MigrateResource(config, &importState, flags.ProjectPath, &migrateResource)
	if len(migrateErr) > 0 {
		var errMessages []string
		for _, e := range migrateErr {
			errMessages = append(errMessages, e.Error())
		}

		return errors.New(strings.Join(errMessages, ","))
	}

	return nil
}

func runApplyCompareTable(supabaseTable []objects.Table, appTable []objects.Table) (migratedData []MigrateItem[objects.Table, objects.UpdateTableItem], err error) {
	result, e := CompareTables(appTable, supabaseTable, CompareModeApply)
	if e != nil {
		err = e
		return
	}

	for i := range result {
		r := result[i]

		migrateType := MigrateTypeIgnore
		if r.IsConflict {
			migrateType = MigrateTypeUpdate
		}

		migratedData = append(migratedData, MigrateItem[objects.Table, objects.UpdateTableItem]{
			Type:           migrateType,
			NewData:        r.SourceResource,
			OldData:        r.TargetResource,
			MigrationItems: r.DiffItems,
		})
	}

	return
}

func runApplyCompareRoles(supabaseRoles []objects.Role, appRoles []objects.Role) error {
	return nil
}

func runApplyCompareRpcFunctions(supabaseFn []objects.Function, appFn []objects.Function) error {
	return nil
}
