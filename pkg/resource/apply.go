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
	"github.com/sev-2/raiden/pkg/resource/types"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase"
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
	Types    []types.MigrateItem
}

type applyDeps struct {
	loadNativeRoles     func() (map[string]raiden.Role, error)
	loadState           func() (*state.State, error)
	extractApp          func(*Flags, *state.State) (state.ExtractTableResult, state.ExtractRoleResult, state.ExtractRpcResult, state.ExtractStorageResult, state.ExtractTypeResult, error)
	loadRemote          func(*Flags, *raiden.Config) (*Resource, error)
	migrate             func(*raiden.Config, *state.LocalState, string, *MigrateData) []error
	buildRoleMigrate    func(state.ExtractRoleResult, []objects.Role) ([]roles.MigrateItem, error)
	buildTableMigrate   func(state.ExtractTableResult, []objects.Table, []string) ([]tables.MigrateItem, error)
	buildRpcMigrate     func(state.ExtractRpcResult, []objects.Function) ([]rpc.MigrateItem, error)
	buildStorageMigrate func(state.ExtractStorageResult, []objects.Bucket) ([]storages.MigrateItem, error)
	buildPolicyMigrate  func(state.ExtractPolicyResult, []objects.Policy) ([]policies.MigrateItem, error)
	buildTypeMigrate    func(state.ExtractTypeResult, []objects.Type) ([]types.MigrateItem, error)
	printReport         func(MigrateData)
}

var defaultApplyDeps = applyDeps{
	loadNativeRoles:     loadMapNativeRole,
	loadState:           state.Load,
	extractApp:          extractAppResource,
	loadRemote:          Load,
	migrate:             Migrate,
	buildRoleMigrate:    roles.BuildMigrateData,
	buildTableMigrate:   tables.BuildMigrateData,
	buildRpcMigrate:     rpc.BuildMigrateData,
	buildStorageMigrate: storages.BuildMigrateData,
	buildPolicyMigrate:  policies.BuildMigrateData,
	buildTypeMigrate:    types.BuildMigrateData,
	printReport:         PrintApplyChangeReport,
}

// applyJob encapsulates the sequential steps required to build and run a migration plan.
// Each helper focuses on a single concern so we can stub them in tests.

type applyJob struct {
	flags            *Flags
	config           *raiden.Config
	deps             applyDeps
	mapNativeRole    map[string]raiden.Role
	latestLocalState *state.State
	localState       state.LocalState
	appTables        state.ExtractTableResult
	appRoles         state.ExtractRoleResult
	appRpcFunctions  state.ExtractRpcResult
	appStorage       state.ExtractStorageResult
	appTypes         state.ExtractTypeResult
	appPolicies      state.ExtractPolicyResult
	resource         *Resource
	migrateData      MigrateData
	reportPrinted    bool
}

// Migrate resource :
//
// [x] migrate table
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
//	[x] add storage acl
//	[x] update storage acl
func Apply(flags *Flags, config *raiden.Config) (err error) {
	return runApply(flags, config, defaultApplyDeps)
}

func runApply(flags *Flags, config *raiden.Config, deps applyDeps) (err error) {
	job := &applyJob{
		flags:  flags,
		config: config,
		deps:   deps,
	}

	defer func() {
		if err == nil && !job.reportPrinted {
			job.printReport()
		}
	}()

	if err = job.run(); err != nil {
		return err
	}
	return nil
}

func (j *applyJob) run() error {
	if j.flags.DryRun {
		ApplyLogger.Info("running apply in dry run mode")
	}

	if err := j.loadNativeRoles(); err != nil {
		return err
	}

	if err := j.loadLocalState(); err != nil {
		return err
	}

	if err := j.extractAppResources(); err != nil {
		return err
	}

	j.buildPoliciesSnapshot()

	if err := j.validateLocalData(); err != nil {
		return err
	}

	if err := j.loadRemoteResource(); err != nil {
		return err
	}

	j.prepareRemoteResource()

	if err := j.buildMigrateData(); err != nil {
		return err
	}

	if err := j.runMigrations(); err != nil {
		return err
	}

	j.printReport()
	return nil
}

func (j *applyJob) loadNativeRoles() error {
	ApplyLogger.Info("load Native log")
	mapNativeRole, err := j.deps.loadNativeRoles()
	if err != nil {
		return err
	}
	j.mapNativeRole = mapNativeRole
	return nil
}

func (j *applyJob) loadLocalState() error {
	ApplyLogger.Info("load resource from local state")
	latestLocalState, err := j.deps.loadState()
	if err != nil {
		return err
	}

	if latestLocalState == nil {
		return errors.New("state file is not found, please run raiden imports first")
	}

	j.latestLocalState = latestLocalState
	j.localState.State = *latestLocalState
	return nil
}

func (j *applyJob) extractAppResources() error {
	ApplyLogger.Info("extract table, role, and rpc from local state")
	appTables, appRoles, appRpcFunctions, appStorage, appTypes, err := j.deps.extractApp(j.flags, j.latestLocalState)
	if err != nil {
		return err
	}

	j.appTables = appTables
	j.appRoles = appRoles
	j.appRpcFunctions = appRpcFunctions
	j.appStorage = appStorage
	j.appTypes = appTypes
	return nil
}

// buildPoliciesSnapshot merges table and storage policies to feed downstream diffing.
func (j *applyJob) buildPoliciesSnapshot() {
	j.appPolicies = mergeAllPolicy(j.appTables, j.appStorage)
}

// validateLocalData checks local relations and policy role references before contacting Supabase.
func (j *applyJob) validateLocalData() error {
	var validateTable []objects.Table
	validateTable = append(validateTable, j.appTables.New.ToFlatTable()...)
	validateTable = append(validateTable, j.appTables.Existing.ToFlatTable()...)
	ApplyLogger.Info("validate local table relation")
	if err := validateTableRelations(validateTable...); err != nil {
		return err
	}

	ApplyLogger.Info("validate local role")
	if err := validateRoleIsExist(j.appPolicies, j.appRoles, j.mapNativeRole); err != nil {
		return err
	}

	return nil
}

func (j *applyJob) loadRemoteResource() error {
	ApplyLogger.Info("load resource from supabase")
	resource, err := j.deps.loadRemote(j.flags, j.config)
	if err != nil {
		return err
	}
	j.resource = resource
	return nil
}

func (j *applyJob) prepareRemoteResource() {
	ApplyLogger.Debug("start filter table and function by allowed schema", "allowed-schema", j.flags.AllowedSchema)
	ApplyLogger.Trace("filter table by schema")
	j.resource.Tables = filterTableBySchema(j.resource.Tables, strings.Split(j.flags.AllowedSchema, ",")...)
	if j.config.Mode == raiden.BffMode && j.config.AllowedTables != "*" {
		allowedTable := strings.Split(j.config.AllowedTables, ",")
		j.resource.Tables = filterAllowedTables(j.resource.Tables, strings.Split(j.flags.AllowedSchema, ","), allowedTable...)
	}
	ApplyLogger.Trace("filter function by schema")
	j.resource.Functions = filterFunctionBySchema(j.resource.Functions, strings.Split(j.flags.AllowedSchema, ",")...)
	ApplyLogger.Debug("finish filter table and function by allowed schema", "allowed-schema", j.flags.AllowedSchema)

	j.resource.Roles = roles.AttachInherithRole(j.mapNativeRole, j.resource.Roles, j.resource.RoleMemberships)
	ApplyLogger.Trace("remove native role for supabase list role")
	j.resource.Roles = filterUserRole(j.resource.Roles, j.mapNativeRole)
}

// buildMigrateData converts extracted state into migrator inputs for each resource type.
func (j *applyJob) buildMigrateData() error {
	ApplyLogger.Info("start build migrate data")

	if j.flags.All() || j.flags.RolesOnly {
		data, err := j.deps.buildRoleMigrate(j.appRoles, j.resource.Roles)
		if err != nil {
			return err
		}
		j.migrateData.Roles = data
	}

	if j.flags.All() || j.flags.ModelsOnly {
		allowedTable := []string{}

		if j.config.Mode == raiden.SvcMode && j.config.AllowedTables != "*" {
			allowedTable = strings.Split(j.config.AllowedTables, ",")
		}

		j.resource.Tables = tables.AttachIndexAndAction(j.resource.Tables, j.resource.Indexes, j.resource.RelationActions)
		data, err := j.deps.buildTableMigrate(j.appTables, j.resource.Tables, allowedTable)
		if err != nil {
			return err
		}
		j.migrateData.Tables = data
	}

	if j.flags.All() || j.flags.RpcOnly {
		data, err := j.deps.buildRpcMigrate(j.appRpcFunctions, j.resource.Functions)
		if err != nil {
			return err
		}
		j.migrateData.Rpc = data
	}

	if j.flags.All() || j.flags.StoragesOnly {
		data, err := j.deps.buildStorageMigrate(j.appStorage, j.resource.Storages)
		if err != nil {
			return err
		}
		j.migrateData.Storages = data
	}

	if len(j.appPolicies.New) > 0 || len(j.appPolicies.Existing) > 0 || len(j.appPolicies.Delete) > 0 {
		data, err := j.deps.buildPolicyMigrate(j.appPolicies, j.resource.Policies)
		if err != nil {
			return err
		}
		j.migrateData.Policies = data
	}

	if len(j.appTypes.New) > 0 || len(j.appTypes.Existing) > 0 || len(j.appTypes.Delete) > 0 {
		data, err := j.deps.buildTypeMigrate(j.appTypes, j.resource.Types)
		if err != nil {
			return err
		}
		j.migrateData.Types = data
	}

	ApplyLogger.Info("finish build migrate data")
	return nil
}

// runMigrations invokes the migrator unless the command is running in dry-run mode.
func (j *applyJob) runMigrations() error {
	if j.flags.DryRun {
		return nil
	}

	errs := j.deps.migrate(j.config, &j.localState, j.flags.ProjectPath, &j.migrateData)
	if len(errs) == 0 {
		ApplyLogger.Info("finish migrate resource")
		return nil
	}

	var errMessages []string
	for _, e := range errs {
		errMessages = append(errMessages, e.Error())
	}

	return errors.New(strings.Join(errMessages, ","))
}

func (j *applyJob) printReport() {
	if j.reportPrinted {
		return
	}
	j.deps.printReport(j.migrateData)
	j.reportPrinted = true
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

	if len(resource.Types) > 0 {
		wg.Add(1)
		go func(w *sync.WaitGroup, eChan chan []error) {
			defer w.Done()

			errors := types.Migrate(config, resource.Types, stateChan, types.ActionFunc)
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

func mergeAllPolicy(et state.ExtractTableResult, es state.ExtractStorageResult) (rs state.ExtractPolicyResult) {
	makeKey := func(p objects.Policy) string {
		return fmt.Sprintf("%s.%s.%s", p.Schema, p.Table, p.Name)
	}

	appendUnique := func(dst *[]objects.Policy, tracker map[string]struct{}, policies ...objects.Policy) {
		for i := range policies {
			policy := policies[i]
			key := makeKey(policy)
			if _, exist := tracker[key]; exist {
				continue
			}

			*dst = append(*dst, policy)
			tracker[key] = struct{}{}
		}
	}

	newTracker := map[string]struct{}{}
	existingTracker := map[string]struct{}{}
	deleteTracker := map[string]struct{}{}

	for i := range et.New {
		t := et.New[i]
		appendUnique(&rs.New, newTracker, t.ExtractedPolicies.New...)
	}

	for i := range et.Existing {
		t := et.Existing[i]
		appendUnique(&rs.New, newTracker, t.ExtractedPolicies.New...)
		appendUnique(&rs.Existing, existingTracker, t.ExtractedPolicies.Existing...)
		appendUnique(&rs.Delete, deleteTracker, t.ExtractedPolicies.Delete...)
	}

	for i := range es.New {
		t := es.New[i]
		appendUnique(&rs.New, newTracker, t.ExtractedPolicies.New...)
	}

	for i := range es.Existing {
		t := es.Existing[i]
		appendUnique(&rs.New, newTracker, t.ExtractedPolicies.New...)
		appendUnique(&rs.Existing, existingTracker, t.ExtractedPolicies.Existing...)
		appendUnique(&rs.Delete, deleteTracker, t.ExtractedPolicies.Delete...)
	}

	for i := range es.Delete {
		t := es.Delete[i]
		appendUnique(&rs.Delete, deleteTracker, t.ExtractedPolicies.New...)
		appendUnique(&rs.Delete, deleteTracker, t.ExtractedPolicies.Existing...)
		appendUnique(&rs.Delete, deleteTracker, t.ExtractedPolicies.Delete...)
	}

	return
}

func validateRoleIsExist(appPolicies state.ExtractPolicyResult, appRoles state.ExtractRoleResult, nativeRole map[string]raiden.Role) error {
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

					if m.NewData.Schema == supabase.DefaultStorageSchema && m.NewData.Table == supabase.DefaultObjectTable {
						fIndex, tState, found := localState.FindStorageByPermissionName(m.NewData.Name)
						if !found {
							continue
						}
						tState.Policies = append(tState.Policies, m.NewData)
						tState.LastUpdate = time.Now()
						localState.UpdateStorage(fIndex, tState)
						ApplyLogger.Trace("new permission", supabase.RlsTypeStorage, tState.Storage.Name, "permission", m.NewData.Name)
					} else {
						fIndex, tState, found := localState.FindTable(m.NewData.TableID)
						if !found {
							continue
						}
						tState.Policies = append(tState.Policies, m.NewData)
						tState.LastUpdate = time.Now()
						ApplyLogger.Trace("new permission", supabase.RlsTypeModel, tState.Table.Name, "permission", m.NewData.Name)
						localState.UpdateTable(fIndex, tState)
					}
				case migrator.MigrateTypeDelete:
					if m.OldData.Name == "" {
						continue
					}

					if m.OldData.Schema == supabase.DefaultStorageSchema && m.OldData.Table == supabase.DefaultObjectTable {
						fIndex, tState, found := localState.FindStorageByPermissionName(m.OldData.Name)
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
							localState.UpdateStorage(fIndex, tState)
							ApplyLogger.Trace("remove permission", supabase.RlsTypeModel, tState.Storage.Name, "permission", m.OldData.Name)
						}
					} else {
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
							ApplyLogger.Trace("remove permission", supabase.RlsTypeModel, tState.Table.Name, "permission", m.OldData.Name)
						}
					}

				case migrator.MigrateTypeUpdate:
					if m.NewData.Name == "" {
						continue
					}
					if m.NewData.Schema == supabase.DefaultStorageSchema && m.NewData.Table == supabase.DefaultObjectTable {
						fIndex, tState, found := localState.FindStorageByPermissionName(m.NewData.Name)
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
							ApplyLogger.Debug("update permission", supabase.RlsTypeStorage, tState.Storage.Name, "permission", m.NewData.Name)
						}
						localState.UpdateStorage(fIndex, tState)
					} else {
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
							ApplyLogger.Warn("new permission",
								supabase.RlsTypeModel, tState.Table.Name,
								"permission", m.NewData.Name,
								"definition", m.NewData.Definition,
								"check", m.NewData.Check,
								"command", m.NewData.Command,
							)
						}
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
						Storage:    m.NewData,
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

					rState.Storage = m.NewData
					rState.LastUpdate = time.Now()
					localState.UpdateStorage(fIndex, rState)
				}
			case *types.MigrateItem:
				switch m.Type {
				case migrator.MigrateTypeCreate:
					if m.NewData.Name == "" {
						continue
					}
					typeStruct := utils.SnakeCaseToPascalCase(m.NewData.Name)
					typePath := fmt.Sprintf("%s/%s/%s.go", projectPath, generator.RoleDir, utils.ToSnakeCase(m.NewData.Name))

					r := state.TypeState{
						Type:       m.NewData,
						TypePath:   typePath,
						TypeStruct: typeStruct,
						LastUpdate: time.Now(),
					}

					localState.AddType(r)
				case migrator.MigrateTypeDelete:
					if m.OldData.Name == "" {
						continue
					}
					localState.DeleteType(m.OldData.ID)
				case migrator.MigrateTypeUpdate:
					fIndex, rState, found := localState.FindType(m.NewData.ID)
					if !found {
						continue
					}

					rState.Type = m.NewData
					rState.LastUpdate = time.Now()
					localState.UpdateType(fIndex, rState)
				}

			}
		}
		done <- localState.Persist()
	}()
	return done
}

func PrintApplyChangeReport(migrateData MigrateData) {
	diffMessage := []string{}

	diffTable := tables.GetDiffChangeMessage(migrateData.Tables)
	if len(diffTable) > 0 {
		diffMessage = append(diffMessage, diffTable)
	}

	diffPolicy := policies.GetDiffChangeMessage(migrateData.Policies)
	if len(diffPolicy) > 0 {
		diffMessage = append(diffMessage, diffPolicy)
	}

	diffRole := roles.GetDiffChangeMessage(migrateData.Roles)
	if len(diffRole) > 0 {
		diffMessage = append(diffMessage, diffRole)
	}

	diffRpc := rpc.GetDiffChangeMessage(migrateData.Rpc)
	if len(diffRpc) > 0 {
		diffMessage = append(diffMessage, diffRpc)
	}

	diffStorage := storages.GetDiffChangeMessage(migrateData.Storages)
	if len(diffStorage) > 0 {
		diffMessage = append(diffMessage, diffStorage)
	}

	diffTypes := types.GetDiffChangeMessage(migrateData.Types)
	if len(diffTypes) > 0 {
		diffMessage = append(diffMessage, diffTypes)
	}

	if len(diffMessage) == 0 {
		ApplyLogger.Info("your code is up to date, nothing to migrate :)")
	} else {
		ApplyLogger.Info("report", "list", strings.Join(diffMessage, "\n"))
	}
}
