package resource

import (
	"errors"
	"fmt"
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
//	[x] create table with acl (rls)
//	[x] delete table (cascade)
//	[x] update table name, schema
//	[x] update table rls enable
//	[x] update table rls force
//	[x] update table with relation - create, update and delete (ordered table by relation)
//	[x] create table with acl (rls), for now all rls outside manage from raiden will be delete
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
// [x] migrate function (Rpc)
func Apply(flags *Flags, config *raiden.Config) error {
	// declare default variable
	var migrateResource MigrateData
	var importState ResourceState

	// load map native role
	logger.Info("apply : load native role")
	mapNativeRole, err := loadMapNativeRole()
	if err != nil {
		return err
	}

	// load app resource
	logger.Info("apply : load resource from local state")
	logger.Debug(strings.Repeat("=", 5), " Load app resource from state")
	latestState, err := loadState()
	if err != nil {
		return err
	}

	if latestState == nil {
		return errors.New("state file is not found, please run raiden imports first")
	} else {
		importState.State = *latestState
	}

	logger.Info("apply : extract table, role, and rpc from local state")
	appTables, appRoles, appRpcFunctions, err := extractAppResource(flags, latestState)
	if err != nil {
		return err
	}
	appPolicies := mergeAllPolicy(appTables)
	logger.Debug(strings.Repeat("=", 5))

	// validate table relation
	var validateTable []objects.Table
	validateTable = append(validateTable, appTables.New.ToFlatTable()...)
	validateTable = append(validateTable, appTables.Existing.ToFlatTable()...)
	logger.Info("apply : validate local table relation")
	if err := validateTableRelations(validateTable...); err != nil {
		return err
	}

	// validate role in policies is exist
	logger.Info("apply : validate local role")
	if err := validateRoleIsExist(appPolicies, appRoles, mapNativeRole); err != nil {
		return err
	}

	// load supabase resource
	logger.Debug(strings.Repeat("=", 5), " Load supabase resource ")
	logger.Info("apply : load table, role and rpc from supabase")
	resource, err := Load(flags, config)
	if err != nil {
		return err
	}
	logger.Debug(strings.Repeat("=", 5))

	// filter table for with allowed schema
	resource.Tables = filterTableBySchema(resource.Tables, strings.Split(flags.AllowedSchema, ",")...)
	resource.Functions = filterFunctionBySchema(resource.Functions, strings.Split(flags.AllowedSchema, ",")...)
	resource.Roles = filterUserRole(resource.Roles, mapNativeRole)

	if flags.All() || flags.RolesOnly {
		logger.Info("apply : compare role")
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
		logger.Info("apply : compare table")
		// bind app table to resource
		if err := bindMigratedTables(appTables, resource.Tables, &migrateResource); err != nil {
			return err
		}

		// bind app policies to resource
		if err := bindMigratedPolicies(appPolicies, resource.Policies, &migrateResource); err != nil {
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

			logger.Debug(strings.Repeat("=", 5), " Table to policies")
			for i := range migrateResource.Policies {
				p := migrateResource.Policies[i]

				var name string
				if p.NewData.Name != "" {
					name = p.NewData.Name
				} else if p.OldData.Name != "" {
					name = p.OldData.Name
				}

				logger.Debugf("- MigrateType : %s, Policy : %s", p.Type, name)
				logger.Debugf(" update items : %+v", p.MigrationItems.ChangeItems)

			}
			logger.Debug(strings.Repeat("=", 5))
		}

	}

	if flags.All() || flags.RpcOnly {
		logger.Info("apply : compare rpc")
		if err := bindMigratedFunctions(appRpcFunctions, resource.Functions, &migrateResource); err != nil {
			return err
		}

		if flags.Verbose {
			logger.Debug(strings.Repeat("=", 5), " Rpc to migrate")
			for i := range migrateResource.Rpc {
				t := migrateResource.Rpc[i]
				var name string
				if t.NewData.Name != "" {
					name = t.NewData.Name
				} else if t.OldData.Name != "" {
					name = t.OldData.Name
				}
				logger.Debugf("- MigrateType : %s, Name: %s", t.Type, name)
			}
			logger.Debug(strings.Repeat("=", 5))
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
	// compare and bind existing table to migrate data
	mapSpRole := make(map[int]bool)
	for i := range spRoles {
		t := spRoles[i]
		mapSpRole[t.ID] = true
	}

	// filter existing table need compare or move to create new
	var compareRoles []objects.Role
	for i := range er.Existing {
		et := er.Existing[i]
		if _, isExist := mapSpRole[et.ID]; isExist {
			compareRoles = append(compareRoles, et)
		} else {
			er.New = append(er.New, et)
		}
	}

	if rs, err := runApplyCompareRoles(spRoles, compareRoles); err != nil {
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

func bindMigratedPolicies(ep state.ExtractedPolicies, spPolicies []objects.Policy, mr *MigrateData) error {
	// compare and bind existing table to migrate data
	mapSpPolicies := make(map[string]bool)
	for i := range spPolicies {
		t := spPolicies[i]
		mapSpPolicies[t.Name] = true
	}

	var comparePolicies []objects.Policy
	for i := range ep.Existing {
		p := ep.Existing[i]
		if _, isExist := mapSpPolicies[p.Name]; isExist {
			comparePolicies = append(comparePolicies, p)
		} else {
			ep.New = append(ep.New, p)
		}
	}

	if rs, err := runApplyComparePolicies(spPolicies, comparePolicies); err != nil {
		return err
	} else {
		mr.Policies = append(mr.Policies, rs...)
	}

	// bind new table to migrated data
	if len(ep.New) > 0 {
		for i := range ep.New {
			t := ep.New[i]
			mr.Policies = append(mr.Policies, MigrateItem[objects.Policy, objects.UpdatePolicyParam]{
				Type:    MigrateTypeCreate,
				NewData: t,
			})
		}
	}

	if len(ep.Delete) > 0 {
		for i := range ep.Delete {
			t := ep.Delete[i]
			isExist := false
			for i := range spPolicies {
				tt := spPolicies[i]
				if tt.Name == t.Name {
					isExist = true
					break
				}
			}

			if isExist {
				mr.Policies = append(mr.Policies, MigrateItem[objects.Policy, objects.UpdatePolicyParam]{
					Type:    MigrateTypeDelete,
					OldData: t,
				})
			}
		}
	}
	return nil
}

func bindMigratedFunctions(er state.ExtractRpcResult, spFn []objects.Function, mr *MigrateData) error {
	if rs, err := runApplyCompareRpcFunctions(spFn, er.Existing); err != nil {
		return err
	} else {
		mr.Rpc = append(mr.Rpc, rs...)
	}

	// bind new table to migrated data
	if len(er.New) > 0 {
		for i := range er.New {
			t := er.New[i]
			mr.Rpc = append(mr.Rpc, MigrateItem[objects.Function, any]{
				Type:    MigrateTypeCreate,
				NewData: t,
			})
		}
	}

	if len(er.Delete) > 0 {
		for i := range er.Delete {
			t := er.Delete[i]
			isExist := false
			for i := range spFn {
				tt := spFn[i]
				if tt.Name == t.Name {
					isExist = true
					break
				}
			}

			if isExist {
				mr.Rpc = append(mr.Rpc, MigrateItem[objects.Function, any]{
					Type:    MigrateTypeDelete,
					OldData: t,
				})
			}
		}
	}

	return nil
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

		r.DiffItems.OldData = r.TargetResource
		migratedData = append(migratedData, MigrateItem[objects.Table, objects.UpdateTableParam]{
			Type:           migrateType,
			NewData:        r.SourceResource,
			OldData:        r.TargetResource,
			MigrationItems: r.DiffItems,
		})
	}

	return
}

func runApplyComparePolicies(supabasePolicies []objects.Policy, appPolicies []objects.Policy) (migratedData []MigrateItem[objects.Policy, objects.UpdatePolicyParam], err error) {
	result := ComparePolicies(appPolicies, supabasePolicies)
	for i := range result {
		r := result[i]
		migrateType := MigrateTypeIgnore
		if r.IsConflict {
			migrateType = MigrateTypeUpdate
		}

		migratedData = append(migratedData, MigrateItem[objects.Policy, objects.UpdatePolicyParam]{
			Type:           migrateType,
			NewData:        r.SourceResource,
			MigrationItems: r.DiffItems,
		})
	}

	return
}

func runApplyCompareRpcFunctions(supabaseFn []objects.Function, appFn []objects.Function) (migratedData []MigrateItem[objects.Function, any], err error) {
	result, e := CompareRpcFunctions(appFn, supabaseFn)
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

		migratedData = append(migratedData, MigrateItem[objects.Function, any]{
			Type:           migrateType,
			NewData:        r.SourceResource,
			OldData:        r.TargetResource,
			MigrationItems: r.DiffItems,
		})
	}

	return
}

func validateTableRelations(migratedTables ...objects.Table) error {
	// convert array data to map data
	mapMigratedTableColumns := make(map[string]bool)
	for i := range migratedTables {
		m := migratedTables[i]

		if m.Name != "" {
			for ic := range m.Columns {
				c := m.Columns[ic]
				key := fmt.Sprintf("%s.%s", m.Name, c.Name)
				mapMigratedTableColumns[key] = true
			}
			continue
		}
	}

	for i := range migratedTables {
		m := migratedTables[i]

		if m.Name == "" {
			return fmt.Errorf("validate relation : invalid table name for create or update")
		}

		for i := range m.Relationships {
			r := m.Relationships[i]
			if r.SourceTableName != m.Name {
				continue
			}

			// validate foreign keys
			fkColumn := fmt.Sprintf("%s.%s", r.SourceTableName, r.SourceColumnName)
			if _, exist := mapMigratedTableColumns[fkColumn]; !exist {
				return fmt.Errorf("validate relation table %s : column %s is not exist in table %s", m.Name, r.SourceColumnName, r.SourceTableName)
			}

			// validate target column
			fkTargetColumn := fmt.Sprintf("%s.%s", r.TargetTableName, r.TargetColumnName)
			if _, exist := mapMigratedTableColumns[fkTargetColumn]; !exist {
				return fmt.Errorf("validate relation table %s : target column %s is not exist in table %s", m.Name, r.TargetColumnName, r.TargetTableName)
			}
		}
	}

	return nil
}

func mergeAllPolicy(et state.ExtractTableResult) (rs state.ExtractedPolicies) {
	for i := range et.New {
		t := et.New[i]
		if len(t.ExtractedPolicies.New) > 0 {
			rs.New = append(rs.New, t.ExtractedPolicies.New...)
		}
	}

	for i := range et.Existing {
		t := et.Existing[i]

		if len(t.ExtractedPolicies.New) > 0 {
			rs.New = append(rs.New, t.ExtractedPolicies.New...)
		}

		if len(t.ExtractedPolicies.Existing) > 0 {
			rs.Existing = append(rs.Existing, t.ExtractedPolicies.Existing...)
		}

		if len(t.ExtractedPolicies.Delete) > 0 {
			rs.Delete = append(rs.Delete, t.ExtractedPolicies.Delete...)
		}
	}

	return
}

func validateRoleIsExist(appPolicies state.ExtractedPolicies, appRoles state.ExtractRoleResult, nativeRole map[string]raiden.Role) error {
	// prepare role data
	var allRoles []objects.Role
	allRoles = append(allRoles, appRoles.New...)
	allRoles = append(allRoles, appRoles.Existing...)

	// merge with native role
	for _, v := range nativeRole {
		r := objects.Role{}
		state.BindToSupabaseRole(&r, v)
		allRoles = append(allRoles, r)
	}

	mapRole := make(map[string]bool)
	for i := range allRoles {
		r := allRoles[i]
		mapRole[r.Name] = true
	}

	// prepare policies data
	var allPolicies []objects.Policy
	allPolicies = append(allPolicies, appPolicies.New...)
	allPolicies = append(allPolicies, appPolicies.Existing...)

	for i := range allPolicies {
		p := allPolicies[i]

		if len(p.Roles) > 0 {
			for _, r := range p.Roles {
				_, exist := mapRole[r]
				if !exist {
					return fmt.Errorf("table %s acl : role %s is not exist", p.Table, r)
				}
			}
		}
	}

	return nil
}
