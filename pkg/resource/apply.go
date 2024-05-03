package resource

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/generator"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/resource/migrator"
	"github.com/sev-2/raiden/pkg/resource/policies"
	"github.com/sev-2/raiden/pkg/resource/roles"
	"github.com/sev-2/raiden/pkg/resource/rpc"
	"github.com/sev-2/raiden/pkg/resource/storages"
	"github.com/sev-2/raiden/pkg/resource/tables"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/utils"
)

var ApplyLogger hclog.Logger = logger.HcLog().Named("apply")

type MigrateData struct {
	Tables   []tables.MigrateItem
	Roles    []roles.MigrateItem
	Rpc      []rpc.MigrateItem
	Policies []policies.MigrateItem
	Storages []storages.MigrateItem
}

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
//
// [x] migrate storage
//
//	[x] create new storage
//	[x] update storage
//	[x] delete storage
//	[ ] add storage acl
//	[ ] update storage acl
func Apply(flags *Flags, config *raiden.Config) error {
	// declare default variable
	var migrateData MigrateData
	var localState state.LocalState

	// load map native role
	ApplyLogger.Info("Load Native log")
	mapNativeRole, err := loadMapNativeRole()
	if err != nil {
		return err
	}

	// load app resource
	ApplyLogger.Info("load resource from local state")
	latestLocalState, err := state.Load()
	if err != nil {
		return err
	}

	if latestLocalState == nil {
		return errors.New("state file is not found, please run raiden imports first")
	} else {
		localState.State = *latestLocalState
	}

	ApplyLogger.Info("extract table, role, and rpc from local state")
	appTables, appRoles, appRpcFunctions, appStorage, err := extractAppResource(flags, latestLocalState)
	if err != nil {
		return err
	}
	appPolicies := mergeAllPolicy(appTables)

	// validate table relation
	var validateTable []objects.Table
	validateTable = append(validateTable, appTables.New.ToFlatTable()...)
	validateTable = append(validateTable, appTables.Existing.ToFlatTable()...)
	ApplyLogger.Info("validate local table relation")
	if err := validateTableRelations(validateTable...); err != nil {
		return err
	}

	// validate role in policies is exist
	ApplyLogger.Info("validate local role")
	if err := validateRoleIsExist(appPolicies, appRoles, mapNativeRole); err != nil {
		return err
	}

	// load supabase resource
	ApplyLogger.Info("load resource from supabase")
	resource, err := Load(flags, config)
	if err != nil {
		return err
	}

	// filter table for with allowed schema
	ApplyLogger.Debug("start - filter table and function by allowed schema", "allowed-schema", flags.AllowedSchema)
	ApplyLogger.Trace("filter table by schema")
	resource.Tables = filterTableBySchema(resource.Tables, strings.Split(flags.AllowedSchema, ",")...)
	ApplyLogger.Trace("filter function by schema")
	resource.Functions = filterFunctionBySchema(resource.Functions, strings.Split(flags.AllowedSchema, ",")...)
	ApplyLogger.Debug("finish - filter table and function by allowed schema", "allowed-schema", flags.AllowedSchema)

	ApplyLogger.Trace("remove native role for supabase list role")
	resource.Roles = filterUserRole(resource.Roles, mapNativeRole)

	ApplyLogger.Info("start build migrate data")
	if flags.All() || flags.RolesOnly {
		if data, err := roles.BuildMigrateData(appRoles, resource.Roles); err != nil {
			return err
		} else {
			migrateData.Roles = data
		}
	}

	if flags.All() || flags.ModelsOnly {
		if data, err := tables.BuildMigrateData(appTables, resource.Tables); err != nil {
			return err
		} else {
			migrateData.Tables = data
		}

		// bind app policies to resource
		if data, err := policies.BuildMigrateData(appPolicies, resource.Policies); err != nil {
			return err
		} else {
			migrateData.Policies = data
		}
	}

	if flags.All() || flags.RpcOnly {
		if data, err := rpc.BuildMigrateData(appRpcFunctions, resource.Functions); err != nil {
			return err
		} else {
			migrateData.Rpc = data
		}
	}

	if flags.All() || flags.StoragesOnly {
		if data, err := storages.BuildMigrateData(appStorage, resource.Storages); err != nil {
			return err
		} else {
			migrateData.Storages = data
		}
	}
	ApplyLogger.Info("finish build migrate data")

	migrateErr := Migrate(config, &localState, flags.ProjectPath, &migrateData)
	if len(migrateErr) > 0 {
		var errMessages []string
		for _, e := range migrateErr {
			errMessages = append(errMessages, e.Error())
		}

		return errors.New(strings.Join(errMessages, ","))
	}
	ApplyLogger.Info("finish migrate resource")
	PrintApplyChangeReport(migrateData)

	return nil
}

func Migrate(config *raiden.Config, importState *state.LocalState, projectPath string, resource *MigrateData) (errors []error) {
	wg, errChan, stateChan := sync.WaitGroup{}, make(chan []error), make(chan any)
	doneListen := UpdateLocalStateFromApply(projectPath, importState, stateChan)

	// role must be run first because will be use when create/update rls
	// and role must be already exist in database
	if len(resource.Roles) > 0 {
		errors = roles.Migrate(config, resource.Roles, stateChan, roles.ActionFunc)
		if len(errors) > 0 {
			close(stateChan)
			return errors
		}
	}

	if len(resource.Tables) > 0 {
		var updateTableRelation []tables.MigrateItem
		for i := range resource.Tables {
			t := resource.Tables[i]
			if len(t.NewData.Relationships) > 0 {
				if t.Type == migrator.MigrateTypeCreate {
					updateTableRelation = append(updateTableRelation, tables.MigrateItem{
						Type:    migrator.MigrateTypeUpdate,
						NewData: t.NewData,
						MigrationItems: objects.UpdateTableParam{
							OldData:             t.NewData,
							ChangeRelationItems: t.MigrationItems.ChangeRelationItems,
							ForceCreateRelation: true,
						},
					})
					resource.Tables[i].MigrationItems.ChangeRelationItems = make([]objects.UpdateRelationItem, 0)
				} else {
					updateTableRelation = append(updateTableRelation, tables.MigrateItem{
						Type:    t.Type,
						NewData: t.NewData,
						OldData: t.OldData,
						MigrationItems: objects.UpdateTableParam{
							OldData:             t.NewData,
							ChangeRelationItems: t.MigrationItems.ChangeRelationItems,
						},
					})
					resource.Tables[i].MigrationItems.ChangeRelationItems = make([]objects.UpdateRelationItem, 0)
				}
			}
		}

		// run migrate for table manipulation
		errors = tables.Migrate(config, resource.Tables, stateChan, tables.ActionFunc)
		if len(errors) > 0 {
			close(stateChan)
			return errors
		}

		// run migrate for relation manipulation
		if len(updateTableRelation) > 0 {
			errors = tables.Migrate(config, updateTableRelation, stateChan, tables.ActionFunc)
			if len(errors) > 0 {
				close(stateChan)
				return errors
			}
		}
	}

	if len(resource.Rpc) > 0 {
		wg.Add(1)
		go func(w *sync.WaitGroup, eChan chan []error) {
			defer wg.Done()

			errors := rpc.Migrate(config, resource.Rpc, stateChan, rpc.ActionFunc)
			if len(errors) > 0 {
				eChan <- errors
				return
			}
		}(&wg, errChan)
	}

	if len(resource.Policies) > 0 {
		wg.Add(1)
		go func(w *sync.WaitGroup, eChan chan []error) {
			defer w.Done()
			errors := policies.Migrate(config, resource.Policies, stateChan, policies.ActionFunc)
			if len(errors) > 0 {
				eChan <- errors
				return
			}
		}(&wg, errChan)
	}

	if len(resource.Storages) > 0 {
		wg.Add(1)
		go func(w *sync.WaitGroup, eChan chan []error) {
			defer w.Done()

			errors := storages.Migrate(config, resource.Storages, stateChan, storages.ActionFunc)
			if len(errors) > 0 {
				eChan <- errors
				return
			}
		}(&wg, errChan)
	}

	go func() {
		wg.Wait()
		close(stateChan)
		close(errChan)
	}()

	for {
		select {
		case rsErr := <-errChan:
			if len(rsErr) > 0 {
				errors = append(errors, rsErr...)
			}
		case saveErr := <-doneListen:
			if saveErr != nil {
				errors = append(errors, saveErr)
			}
			return
		}
	}
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

func UpdateLocalStateFromApply(projectPath string, localState *state.LocalState, stateChan chan any) (done chan error) {
	done = make(chan error)
	go func() {
		for rs := range stateChan {
			if rs == nil {
				continue
			}

			switch m := rs.(type) {
			case *tables.MigrateItem:
				switch m.Type {
				case migrator.MigrateTypeCreate:
					if m.NewData.Name == "" {
						continue
					}
					modelStruct := utils.SnakeCaseToPascalCase(m.NewData.Name)
					modelPath := fmt.Sprintf("%s/%s/%s.go", projectPath, generator.ModelDir, utils.ToSnakeCase(m.NewData.Name))

					ts := state.TableState{
						Table:       m.NewData,
						ModelPath:   modelPath,
						ModelStruct: modelStruct,
						LastUpdate:  time.Now(),
					}

					localState.AddTable(ts)
				case migrator.MigrateTypeDelete:
					if m.OldData.Name == "" {
						continue
					}

					localState.DeleteTable(m.OldData.ID)
				case migrator.MigrateTypeUpdate:
					fIndex, tState, found := localState.FindTable(m.NewData.ID)
					if !found {
						continue
					}

					tState.Table = m.NewData
					tState.LastUpdate = time.Now()
					localState.UpdateTable(fIndex, tState)
				}
			case *roles.MigrateItem:
				switch m.Type {
				case migrator.MigrateTypeCreate:
					if m.NewData.Name == "" {
						continue
					}
					roleStruct := utils.SnakeCaseToPascalCase(m.NewData.Name)
					rolePath := fmt.Sprintf("%s/%s/%s.go", projectPath, generator.RoleDir, utils.ToSnakeCase(m.NewData.Name))

					r := state.RoleState{
						Role:       m.NewData,
						RolePath:   rolePath,
						RoleStruct: roleStruct,
						LastUpdate: time.Now(),
					}

					localState.AddRole(r)
				case migrator.MigrateTypeDelete:
					if m.OldData.Name == "" {
						continue
					}
					localState.DeleteRole(m.OldData.ID)
				case migrator.MigrateTypeUpdate:
					fIndex, rState, found := localState.FindRole(m.NewData.ID)
					if !found {
						continue
					}

					rState.Role = m.NewData
					rState.LastUpdate = time.Now()
					localState.UpdateRole(fIndex, rState)
				}
			case *policies.MigrateItem:
				switch m.Type {
				case migrator.MigrateTypeCreate:
					if m.NewData.Name == "" {
						continue
					}

					fIndex, tState, found := localState.FindTable(m.NewData.TableID)
					if !found {
						continue
					}
					tState.Policies = append(tState.Policies, m.NewData)
					tState.LastUpdate = time.Now()
					localState.UpdateTable(fIndex, tState)
				case migrator.MigrateTypeDelete:
					if m.OldData.Name == "" {
						continue
					}
					fIndex, tState, found := localState.FindTable(m.OldData.TableID)
					if !found {
						continue
					}

					// find policy index
					pi := -1
					for i := range tState.Policies {
						p := tState.Policies[i]
						if p.ID == m.OldData.ID {
							pi = i
							break
						}
					}

					if pi > -1 {
						tState.Policies = append(tState.Policies[:pi], tState.Policies[pi+1:]...)
						tState.LastUpdate = time.Now()
						localState.UpdateTable(fIndex, tState)
					}
				case migrator.MigrateTypeUpdate:
					if m.NewData.Name == "" {
						continue
					}
					fIndex, tState, found := localState.FindTable(m.NewData.TableID)
					if !found {
						continue
					}

					// find policy index
					pi := -1
					for i := range tState.Policies {
						p := tState.Policies[i]
						if p.ID == m.OldData.ID {
							pi = i
							break
						}
					}
					if pi > -1 {
						tState.Policies[pi] = m.NewData
						tState.LastUpdate = time.Now()
						localState.UpdateTable(fIndex, tState)
					}
				}
			case *rpc.MigrateItem:
				switch m.Type {
				case migrator.MigrateTypeCreate:
					if m.NewData.Name == "" {
						continue
					}
					rpcStruct := utils.SnakeCaseToPascalCase(m.NewData.Name)
					rpcPath := fmt.Sprintf("%s/%s/%s.go", projectPath, generator.RpcDir, utils.ToSnakeCase(m.NewData.Name))

					r := state.RpcState{
						Function:   m.NewData,
						RpcPath:    rpcPath,
						RpcStruct:  rpcStruct,
						LastUpdate: time.Now(),
					}
					localState.AddRpc(r)
				case migrator.MigrateTypeDelete:
					if m.OldData.Name == "" {
						continue
					}
					localState.DeleteRpc(m.OldData.ID)
				case migrator.MigrateTypeUpdate:
					fIndex, rState, found := localState.FindRpc(m.NewData.ID)
					if !found {
						continue
					}

					rState.Function = m.NewData
					rState.LastUpdate = time.Now()
					localState.UpdateRpc(fIndex, rState)
				}
			case *storages.MigrateItem:
				switch m.Type {
				case migrator.MigrateTypeCreate:
					if m.NewData.Name == "" {
						continue
					}

					r := state.StorageState{
						Bucket:     m.NewData,
						LastUpdate: time.Now(),
					}

					localState.AddStorage(r)
				case migrator.MigrateTypeDelete:
					if m.OldData.Name == "" {
						continue
					}
					localState.DeleteStorage(m.OldData.ID)
				case migrator.MigrateTypeUpdate:
					fIndex, rState, found := localState.FindStorage(m.NewData.ID)
					if !found {
						continue
					}

					rState.Bucket = m.NewData
					rState.LastUpdate = time.Now()
					localState.UpdateStorage(fIndex, rState)
				}
			}
		}
		done <- localState.Persist()
	}()
	return done
}

func PrintApplyChangeReport(migrateData MigrateData) {
	diffTable := tables.GetDiffChangeMessage(migrateData.Tables)
	diffPolicy := policies.GetDiffChangeMessage(migrateData.Policies)
	diffRole := roles.GetDiffChangeMessage(migrateData.Roles)
	diffRpc := rpc.GetDiffChangeMessage(migrateData.Rpc)
	diffStorage := storages.GetDiffChangeMessage(migrateData.Storages)
	diffMessage := []string{
		diffTable, diffPolicy, diffRole, diffRpc, diffStorage,
	}
	ApplyLogger.Info("report", "list", strings.Join(diffMessage, "\n"))
}
