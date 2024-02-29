package resource

import (
	"errors"
	"strings"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase/objects"
)

// Migrate resource :
//
// [ ] migrate table
//
//	[x] create table name, schema and columns
//	[x] create table rls enable
//	[x] create table rls force
//	[x] create table with primary key
//	[x] create table with relation (ordered table by relation)
//	[ ] create table with acl (rls)
//	[x] delete table
//	[x] update table name, schema
//	[x] update table rls enable
//	[x] update table rls force
//	[x] update table with relation - create, update and delete (ordered table by relation)
//	[ ] create table with acl (rls)
//	[x] update table column add new column
//	[x] update table column delete column
//	[x] update table column set default schema
//	[x] update table column set set data type
//	[x] update table column set unique column
//	[x] update table column set nullable
//
// [ ] migrate role
//
//	[x] create new role
//	[x] create role with connection limit
//	[x] create role with inherit
//	[ ] create role with is replicate - unsupported now because need superuser role
//	[ ] create role with is super user - unsupported now because need superuser role
//	[x] create role with can bypass rls, can create db, can create role, can login
//	[x] create role with valid until
//	[x] delete role
//	[ ] update name
//	[x] update connection limit
//	[x] update inherit
//	[ ] update is replicate
//	[ ] update is super user
//	[x] update can bypass rls
//	[x] update can create db
//	[x] update can create role
//	[x] update can login
//	[x] update valid until
//
// [ ] migrate function
func Apply(flags *Flags, config *raiden.Config) error {
	// declare default variable
	var migrateResource MigrateData
	var importState ResourceState

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
	latestState, err := loadState()
	if err != nil {
		return err
	}

	if latestState == nil {
		return errors.New("state file is not found, please run raiden import first")
	}

	importState.State = *latestState

	appTables, appRoles, appRpcFunctions, err := extractAppResource(flags, latestState)
	if err != nil {
		return err
	}
	logger.Debug(strings.Repeat("=", 5))

	if flags.All() || flags.RolesOnly {
		if err := bindMigratedRoles(appRoles, resource.Roles, &migrateResource); err != nil {
			return err
		}

		if flags.Verbose {
			logger.Debug(strings.Repeat("=", 5), " Role to migrate")
			for i := range migrateResource.Roles {
				t := migrateResource.Roles[i]
				var name string
				if t.NewData.Name != "" {
					name = t.NewData.Name
				} else if t.OldData.Name != "" {
					name = t.OldData.Name
				}
				logger.Debugf("- MigrateType : %s, Name: %s , update items : %+v", t.Type, name, t.MigrationItems.ChangeItems)
				logger.Debugf(" update items : %+v", t.MigrationItems.ChangeItems)
			}
			logger.Debug(strings.Repeat("=", 5))
		}
	}

	if flags.All() || flags.ModelsOnly {
		// bind app table to resource
		if err := bindMigratedTables(appTables, resource.Tables, &migrateResource); err != nil {
			return err
		}

		if flags.Verbose {
			logger.Debug(strings.Repeat("=", 5), " Table to migrate")
			for i := range migrateResource.Tables {
				t := migrateResource.Tables[i]

				var name string
				if t.NewData.Name != "" {
					name = t.NewData.Name
				} else if t.OldData.Name != "" {
					name = t.OldData.Name
				}

				logger.Debugf("- MigrateType : %s, Table : %s", t.Type, name)
				logger.Debugf(" update items : %+v", t.MigrationItems.ChangeItems)
				logger.Debugf(" update column  : %+v", t.MigrationItems.ChangeColumnItems)
				logger.Debugf(" update relations  : %+v", t.MigrationItems.ChangeRelationItems)

			}
			logger.Debug(strings.Repeat("=", 5))
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

func bindMigratedTables(etr state.ExtractTableResult, spTables []objects.Table, mr *MigrateData) error {
	// compare and bind existing table to migrate data
	mapSpTable := make(map[int]bool)
	for i := range spTables {
		t := spTables[i]
		mapSpTable[t.ID] = true
	}

	// filter existing table need compare or move to create new
	var compareTables []objects.Table
	for i := range etr.Existing {
		et := etr.Existing[i]
		if _, isExist := mapSpTable[et.Table.ID]; isExist {
			compareTables = append(compareTables, et.Table)
		} else {
			etr.New = append(etr.New, et)
		}
	}
	if rs, err := runApplyCompareTable(spTables, compareTables); err != nil {
		return err
	} else {
		mr.Tables = append(mr.Tables, rs...)
	}

	// bind new table to migrated data
	if len(etr.New) > 0 {
		for i := range etr.New {
			t := etr.New[i]
			mr.Tables = append(mr.Tables, MigrateItem[objects.Table, objects.UpdateTableParam]{
				Type:    MigrateTypeCreate,
				NewData: t.Table,
			})

			if len(t.Policies) > 0 {
				for ip := range t.Policies {
					p := t.Policies[ip]
					mr.Policies = append(mr.Policies, MigrateItem[objects.Policy, any]{
						Type:    MigrateTypeCreate,
						NewData: p,
					})
				}

			}
		}
	}

	// bind delete table to migrate data
	if len(etr.Delete) > 0 {
		for i := range etr.Delete {
			t := etr.Delete[i]
			isExist := false
			for i := range spTables {
				tt := spTables[i]
				if tt.Name == t.Table.Name {
					isExist = true
					break
				}
			}

			if isExist {
				mr.Tables = append(mr.Tables, MigrateItem[objects.Table, objects.UpdateTableParam]{
					Type:    MigrateTypeDelete,
					OldData: t.Table,
				})
			}
		}
	}

	return nil
}

func bindMigratedRoles(er state.ExtractRoleResult, spRoles []objects.Role, mr *MigrateData) error {
	if rs, err := runApplyCompareRoles(spRoles, er.Existing); err != nil {
		return err
	} else {
		mr.Roles = append(mr.Roles, rs...)
	}

	// bind new table to migrated data
	if len(er.New) > 0 {
		for i := range er.New {
			t := er.New[i]
			mr.Roles = append(mr.Roles, MigrateItem[objects.Role, objects.UpdateRoleParam]{
				Type:    MigrateTypeCreate,
				NewData: t,
			})
		}
	}

	if len(er.Delete) > 0 {
		for i := range er.Delete {
			t := er.Delete[i]
			isExist := false
			for i := range spRoles {
				tt := spRoles[i]
				if tt.Name == t.Name {
					isExist = true
					break
				}
			}

			if isExist {
				mr.Roles = append(mr.Roles, MigrateItem[objects.Role, objects.UpdateRoleParam]{
					Type:    MigrateTypeDelete,
					OldData: t,
				})
			}
		}
	}

	return nil
}

func runApplyCompareTable(supabaseTable []objects.Table, appTable []objects.Table) (migratedData []MigrateItem[objects.Table, objects.UpdateTableParam], err error) {
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

		r.DiffItems.OldData.Name = r.Name
		migratedData = append(migratedData, MigrateItem[objects.Table, objects.UpdateTableParam]{
			Type:           migrateType,
			NewData:        r.SourceResource,
			OldData:        r.TargetResource,
			MigrationItems: r.DiffItems,
		})
	}

	return
}

func runApplyCompareRoles(supabaseRoles []objects.Role, appRoles []objects.Role) (migratedData []MigrateItem[objects.Role, objects.UpdateRoleParam], err error) {
	result, e := CompareRoles(appRoles, supabaseRoles, CompareModeApply)
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

		r.DiffItems.OldData = r.TargetResource
		migratedData = append(migratedData, MigrateItem[objects.Role, objects.UpdateRoleParam]{
			Type:           migrateType,
			NewData:        r.SourceResource,
			OldData:        r.TargetResource,
			MigrationItems: r.DiffItems,
		})
	}

	return
}

func runApplyCompareRpcFunctions(supabaseFn []objects.Function, appFn []objects.Function) error {
	return nil
}
