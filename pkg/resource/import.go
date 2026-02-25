package resource

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/generator"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/resource/roles"
	"github.com/sev-2/raiden/pkg/resource/rpc"
	"github.com/sev-2/raiden/pkg/resource/storages"
	"github.com/sev-2/raiden/pkg/resource/tables"
	"github.com/sev-2/raiden/pkg/resource/types"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/utils"
)

var ImportLogger hclog.Logger = logger.HcLog().Named("import")

type importDeps struct {
	loadNativeRoles func() (map[string]raiden.Role, error)
	loadRemote      func(*Flags, *raiden.Config) (*Resource, error)
	loadState       func() (*state.State, error)
	extractApp      func(*Flags, *state.State) (state.ExtractTableResult, state.ExtractRoleResult, state.ExtractRpcResult, state.ExtractStorageResult, state.ExtractTypeResult, error)
	compareTypes    func([]objects.Type, []objects.Type) error
	compareTables   func([]objects.Table, []objects.Table) error
	compareRoles    func([]objects.Role, []objects.Role) error
	compareRpc      func([]objects.Function, []objects.Function) error
	compareStorages func([]objects.Bucket, []objects.Bucket) error
	updateStateOnly func(*state.LocalState, *Resource, map[string]state.ModelValidationTag) error
	generate        func(*raiden.Config, *state.LocalState, string, *Resource, map[string]state.ModelValidationTag, bool) error
	printReport     func(ImportReport, bool)
}

var defaultImportDeps = importDeps{
	loadNativeRoles: loadMapNativeRole,
	loadRemote:      Load,
	loadState:       state.Load,
	extractApp:      extractAppResource,
	compareTypes: func(remote []objects.Type, existing []objects.Type) error {
		return types.Compare(remote, existing)
	},
	compareTables: func(remote []objects.Table, existing []objects.Table) error {
		return tables.Compare(tables.CompareModeImport, remote, existing)
	},
	compareRoles:    roles.Compare,
	compareRpc:      rpc.Compare,
	compareStorages: storages.Compare,
	updateStateOnly: updateStateOnly,
	generate:        generateImportResource,
	printReport:     PrintImportReport,
}

// importJob keeps the shared state and dependencies needed to process an import.
// The methods on this struct execute individual phases so tests can stub them easily.

type importJob struct {
	flags                  *Flags
	config                 *raiden.Config
	deps                   importDeps
	mapNativeRole          map[string]raiden.Role
	resource               *Resource
	localState             *state.State
	importState            state.LocalState
	appTables              state.ExtractTableResult
	appRoles               state.ExtractRoleResult
	appRpcFunctions        state.ExtractRpcResult
	appStorage             state.ExtractStorageResult
	appTypes               state.ExtractTypeResult
	nativeStateRoles       []state.RoleState
	dryRunErrors           []string
	mapModelValidationTags map[string]state.ModelValidationTag
	report                 ImportReport
	reportComputed         bool
	reportPrinted          bool
	skipReport             bool
}

// List of import resource
// [x] import table, relation, column specification and acl
// [x] import role
// [x] import function
// [x] import storage
// [x] import policy
func Import(flags *Flags, config *raiden.Config) (err error) {
	return runImport(flags, config, defaultImportDeps)
}

// runImport wires the provided dependencies into an importJob and executes the workflow.
func runImport(flags *Flags, config *raiden.Config, deps importDeps) (err error) {
	job := &importJob{
		flags:                  flags,
		config:                 config,
		deps:                   deps,
		dryRunErrors:           []string{},
		mapModelValidationTags: make(map[string]state.ModelValidationTag),
	}

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("import panic: %v", r)
		}
		if err == nil && job.reportComputed && !job.reportPrinted && !job.skipReport {
			job.printReport(flags.DryRun)
		}
	}()

	if err = job.run(); err != nil {
		return err
	}
	return nil
}

func (j *importJob) run() error {
	if j.flags.DryRun {
		ImportLogger.Info("running import in dry run mode")
	}

	if err := j.loadNativeRoles(); err != nil {
		return err
	}

	if err := j.loadRemoteResource(); err != nil {
		return err
	}

	j.prepareRemoteResource()

	if err := j.loadLocalState(); err != nil {
		return err
	}

	if err := j.extractAppResources(); err != nil {
		return err
	}

	j.collectValidationTags()

	if err := j.performComparisons(); err != nil {
		return err
	}

	j.computeReport()

	return j.handleOutput()
}

func (j *importJob) loadNativeRoles() error {
	ImportLogger.Info("load native role")
	mapNativeRole, err := j.deps.loadNativeRoles()
	if err != nil {
		return err
	}
	j.mapNativeRole = mapNativeRole
	return nil
}

func (j *importJob) loadRemoteResource() error {
	ImportLogger.Info("load resource from database")
	resource, err := j.deps.loadRemote(j.flags, j.config)
	if err != nil {
		return err
	}

	resource.Tables = tables.AttachIndexAndAction(resource.Tables, resource.Indexes, resource.RelationActions)
	resource.Roles = roles.AttachInherithRole(j.mapNativeRole, resource.Roles, resource.RoleMemberships)
	j.resource = resource
	return nil
}

func (j *importJob) prepareRemoteResource() {
	ImportLogger.Debug("get native roles")
	j.nativeStateRoles = filterIsNativeRole(j.mapNativeRole, j.resource.Roles)

	ImportLogger.Debug("start filter table and function by allowed schema", "allowed-schema", j.flags.AllowedSchema)
	ImportLogger.Trace("filter table by schema")

	j.resource.Tables = filterTableBySchema(j.resource.Tables, strings.Split(j.flags.AllowedSchema, ",")...)

	if j.config.Mode == raiden.BffMode && j.config.AllowedTables != "*" {
		allowedTable := strings.Split(j.config.AllowedTables, ",")
		j.resource.Tables = filterAllowedTables(j.resource.Tables, strings.Split(j.flags.AllowedSchema, ","), allowedTable...)
	}

	// log tables included and detect relations referencing missing tables
	importedTableSet := make(map[string]bool)
	for _, t := range j.resource.Tables {
		importedTableSet[fmt.Sprintf("%s.%s", t.Schema, t.Name)] = true
	}
	ImportLogger.Debug("tables included for import", "count", len(j.resource.Tables))
	for _, t := range j.resource.Tables {
		ImportLogger.Trace("included table", "schema", t.Schema, "table", t.Name, "relations", len(t.Relationships))
		for _, r := range t.Relationships {
			targetKey := fmt.Sprintf("%s.%s", r.TargetTableSchema, r.TargetTableName)
			sourceKey := fmt.Sprintf("%s.%s", r.SourceSchema, r.SourceTableName)
			if !importedTableSet[targetKey] {
				ImportLogger.Debug("relation target table not in import set", "table", t.Name, "relation-target", r.TargetTableName, "target-schema", r.TargetTableSchema)
			}
			if !importedTableSet[sourceKey] {
				ImportLogger.Debug("relation source table not in import set", "table", t.Name, "relation-source", r.SourceTableName, "source-schema", r.SourceSchema)
			}
		}
	}

	ImportLogger.Trace("filter function by schema")
	j.resource.Functions = filterFunctionBySchema(j.resource.Functions, strings.Split(j.flags.AllowedSchema, ",")...)
	ImportLogger.Debug("finish filter table and function by allowed schema")

	ImportLogger.Trace("remove native role for supabase list role")
	j.resource.Roles = filterUserRole(j.resource.Roles, j.mapNativeRole)
}

func (j *importJob) loadLocalState() error {
	ImportLogger.Info("load resource from local state")
	localState, err := j.deps.loadState()
	if err != nil {
		return err
	}
	j.localState = localState
	return nil
}

func (j *importJob) extractAppResources() error {
	ImportLogger.Info("extract data from local state")
	appTables, appRoles, appRpcFunctions, appStorage, appType, err := j.deps.extractApp(j.flags, j.localState)
	if err != nil {
		return err
	}

	j.appTables = appTables
	j.appRoles = appRoles
	j.appRpcFunctions = appRpcFunctions
	j.appStorage = appStorage
	j.appTypes = appType

	j.importState = state.LocalState{
		State: state.State{
			Roles: j.nativeStateRoles,
		},
	}

	if j.flags.ForceImport {
		ImportLogger.Warn("force import enabled: skipping diff checks and overwriting local state")
	}

	return nil
}

// collectValidationTags records existing model validation tags so regeneration preserves them.
func (j *importJob) collectValidationTags() {
	if !(j.flags.All() || j.flags.ModelsOnly) {
		return
	}

	for i := range j.appTables.New {
		nt := j.appTables.New[i]
		if nt.ValidationTags != nil {
			j.mapModelValidationTags[nt.Table.Name] = nt.ValidationTags
		}
	}

	for i := range j.appTables.Existing {
		et := j.appTables.Existing[i]
		if et.ValidationTags != nil {
			j.mapModelValidationTags[et.Table.Name] = et.ValidationTags
		}
	}
}

// performComparisons runs diff checks against the remote Supabase state unless forced import is requested.
func (j *importJob) performComparisons() error {
	if j.flags.ForceImport {
		ImportLogger.Info("skip diff comparison and overwrite local state")
		return nil
	}

	ImportLogger.Info("compare supabase resource and local resource")

	if err := j.compareTypes(); err != nil {
		return err
	}
	if err := j.compareTables(); err != nil {
		return err
	}
	if err := j.compareRoles(); err != nil {
		return err
	}
	if err := j.compareRpc(); err != nil {
		return err
	}
	if err := j.compareStorages(); err != nil {
		return err
	}
	return nil
}

func (j *importJob) compareTypes() error {
	if !(j.flags.All() || j.flags.ModelsOnly) || len(j.appTypes.Existing) == 0 {
		return nil
	}
	if !j.flags.DryRun {
		ImportLogger.Debug("start compare types")
	}
	if err := j.deps.compareTypes(j.resource.Types, j.appTypes.Existing); err != nil {
		if j.flags.DryRun {
			j.dryRunErrors = append(j.dryRunErrors, err.Error())
			return nil
		}
		return err
	}
	if !j.flags.DryRun {
		ImportLogger.Debug("finish compare types")
	}
	return nil
}

func (j *importJob) compareTables() error {
	if !(j.flags.All() || j.flags.ModelsOnly) || len(j.appTables.Existing) == 0 {
		return nil
	}
	if !j.flags.DryRun {
		ImportLogger.Debug("start compare table")
	}
	compareTables := make([]objects.Table, 0, len(j.appTables.Existing))
	for i := range j.appTables.Existing {
		compareTables = append(compareTables, j.appTables.Existing[i].Table)
	}

	// Build set of locally-known tables so we can filter out remote
	// relationships that reference tables not in the local model set
	// (e.g., cross-schema FKs to auth.users). The generator already
	// skips these when creating join tags, so the comparison should too.
	localTableSet := make(map[string]bool, len(compareTables))
	for _, t := range compareTables {
		localTableSet[fmt.Sprintf("%s.%s", t.Schema, t.Name)] = true
	}
	remoteTables := make([]objects.Table, len(j.resource.Tables))
	copy(remoteTables, j.resource.Tables)
	for i := range remoteTables {
		if len(remoteTables[i].Relationships) == 0 {
			continue
		}
		filtered := make([]objects.TablesRelationship, 0, len(remoteTables[i].Relationships))
		for _, r := range remoteTables[i].Relationships {
			targetKey := fmt.Sprintf("%s.%s", r.TargetTableSchema, r.TargetTableName)
			sourceKey := fmt.Sprintf("%s.%s", r.SourceSchema, r.SourceTableName)
			if localTableSet[targetKey] && localTableSet[sourceKey] {
				filtered = append(filtered, r)
			}
		}
		remoteTables[i].Relationships = filtered
	}

	if err := j.deps.compareTables(remoteTables, compareTables); err != nil {
		if j.flags.DryRun {
			j.dryRunErrors = append(j.dryRunErrors, err.Error())
			return nil
		}
		return err
	}

	if !j.flags.DryRun {
		ImportLogger.Debug("finish compare table")
	}

	return nil
}

func (j *importJob) compareRoles() error {
	if !(j.flags.All() || j.flags.RolesOnly) || len(j.appRoles.Existing) == 0 {
		return nil
	}
	if !j.flags.DryRun {
		ImportLogger.Debug("start compare role")
	}
	if err := j.deps.compareRoles(j.resource.Roles, j.appRoles.Existing); err != nil {
		if j.flags.DryRun {
			j.dryRunErrors = append(j.dryRunErrors, err.Error())
			return nil
		}
		return err
	}
	if !j.flags.DryRun {
		ImportLogger.Debug("finish compare role")
	}
	return nil
}

func (j *importJob) compareRpc() error {
	if !(j.flags.All() || j.flags.RpcOnly) || len(j.appRpcFunctions.Existing) == 0 {
		return nil
	}
	if !j.flags.DryRun {
		ImportLogger.Debug("start compare rpc")
	}

	// Restore state CompleteStatement for import comparison.
	// BindRpcFunction rebuilds CompleteStatement from the Go struct template which
	// may differ in formatting from pg_get_functiondef() (param prefix, default
	// quoting, search_path inclusion). Using the stored state value (captured from
	// the last import) ensures we only flag real remote changes as conflicts.
	mapStateCS := make(map[string]string)
	for _, rs := range j.localState.Rpc {
		if rs.Function.CompleteStatement != "" {
			mapStateCS[rs.Function.Name] = rs.Function.CompleteStatement
		}
	}
	for i := range j.appRpcFunctions.Existing {
		if cs, ok := mapStateCS[j.appRpcFunctions.Existing[i].Name]; ok {
			j.appRpcFunctions.Existing[i].CompleteStatement = cs
		}
	}

	if err := j.deps.compareRpc(j.resource.Functions, j.appRpcFunctions.Existing); err != nil {
		if j.flags.DryRun {
			j.dryRunErrors = append(j.dryRunErrors, err.Error())
			return nil
		}
		return err
	}
	if !j.flags.DryRun {
		ImportLogger.Debug("finish compare rpc")
	}
	return nil
}

func (j *importJob) compareStorages() error {
	if !(j.flags.All() || j.flags.StoragesOnly) || len(j.appStorage.Existing) == 0 {
		return nil
	}
	if !j.flags.DryRun {
		ImportLogger.Debug("start compare storage")
	}
	compareStorages := make([]objects.Bucket, 0, len(j.appStorage.Existing))
	for i := range j.appStorage.Existing {
		compareStorages = append(compareStorages, j.appStorage.Existing[i].Storage)
	}

	if err := j.deps.compareStorages(j.resource.Storages, compareStorages); err != nil {
		if j.flags.DryRun {
			j.dryRunErrors = append(j.dryRunErrors, err.Error())
			return nil
		}
		return err
	}
	if !j.flags.DryRun {
		ImportLogger.Debug("finish compare storage")
	}
	return nil
}

func (j *importJob) computeReport() {
	j.report = ImportReport{
		Role:    roles.GetNewCountData(j.resource.Roles, j.appRoles),
		Table:   tables.GetNewCountData(j.resource.Tables, j.appTables),
		Storage: storages.GetNewCountData(j.resource.Storages, j.appStorage),
		Rpc:     rpc.GetNewCountData(j.resource.Functions, j.appRpcFunctions),
		Types:   types.GetNewCountData(j.resource.Types, j.appTypes),
	}
	j.reportComputed = true
}

// handleOutput decides whether to mutate local files or just display the report based on flags.
func (j *importJob) handleOutput() error {
	if !j.flags.DryRun {
		if j.flags.UpdateStateOnly {
			return j.deps.updateStateOnly(&j.importState, j.resource, j.mapModelValidationTags)
		}

		if err := j.deps.generate(j.config, &j.importState, j.flags.ProjectPath, j.resource, j.mapModelValidationTags, j.flags.GenerateController); err != nil {
			return err
		}
		j.printReport(false)
		return nil
	}

	if len(j.dryRunErrors) > 0 {
		errMessage := strings.Join(j.dryRunErrors, "\n")
		ImportLogger.Error("got error", "err-msg", errMessage)
		j.skipReport = true
		return nil
	}

	j.printReport(true)
	return nil
}

func (j *importJob) printReport(dryRun bool) {
	if j.reportPrinted {
		return
	}
	j.deps.printReport(j.report, dryRun)
	j.reportPrinted = true
}

// ----- Generate import data -----
func generateImportResource(config *raiden.Config, importState *state.LocalState, projectPath string, resource *Resource, mapModelValidationTags map[string]state.ModelValidationTag, generateController bool) error {
	if err := generator.CreateInternalFolder(projectPath); err != nil {
		return err
	}

	wg, errChan, stateChan := sync.WaitGroup{}, make(chan error), make(chan any)
	doneListen := UpdateLocalStateFromImport(importState, stateChan)
	roleMap := make(map[string]string)
	nativeRoleMap := make(map[string]raiden.Role)

	if len(resource.Tables) > 0 || len(resource.Storages) > 0 {
		if len(resource.Roles) > 0 {
			for _, i := range resource.Roles {
				roleMap[i.Name] = i.Name
			}
		}
		nativeRoleMap, _ = loadMapNativeRole()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		if len(resource.Types) > 0 {
			ImportLogger.Info("start generate types")
			captureFunc := ImportDecorateFunc(resource.Types, func(item objects.Type, input generator.GenerateInput) bool {
				if i, ok := input.BindData.(generator.GenerateTypeData); ok {
					if i.Name == item.Name {
						return true
					}
				}
				return false
			}, stateChan)

			if err := generator.GenerateTypes(projectPath, resource.Types, captureFunc); err != nil {
				errChan <- err
			}
			ImportLogger.Info("finish generate types")
		}

		if len(resource.Tables) > 0 {
			tableInputs := tables.BuildGenerateModelInputs(resource.Tables, resource.Policies, mapModelValidationTags)
			ImportLogger.Info("start generate tables")
			captureFunc := ImportDecorateFunc(tableInputs, func(item *generator.GenerateModelInput, input generator.GenerateInput) bool {
				if i, ok := input.BindData.(generator.GenerateModelData); ok {
					if i.StructName == utils.SnakeCaseToPascalCase(item.Table.Name) {
						return true
					}
				}
				return false
			}, stateChan)

			var mapDataType = make(map[string]objects.Type)
			for i := range resource.Types {
				dataType := resource.Types[i]
				mapDataType[dataType.Name] = dataType
			}

			if err := generator.GenerateModels(projectPath, config.ProjectName, tableInputs, mapDataType, roleMap, nativeRoleMap, captureFunc); err != nil {
				errChan <- err
			}

			if generateController && config.Mode == raiden.BffMode {
				if err := generator.GenerateRestControllers(projectPath, config.ProjectName, tableInputs, generator.Generate); err != nil {
					errChan <- err
				}
			}

			ImportLogger.Info("finish generate tables")
		}

		// generate all roles from cloud / pg-meta
		if len(resource.Roles) > 0 {
			ImportLogger.Info("start generate roles")
			captureFunc := ImportDecorateFunc(resource.Roles, func(item objects.Role, input generator.GenerateInput) bool {
				if i, ok := input.BindData.(generator.GenerateRoleData); ok {
					if i.Name == item.Name {
						return true
					}
				}
				return false
			}, stateChan)

			if err := generator.GenerateRoles(projectPath, resource.Roles, captureFunc); err != nil {
				errChan <- err
			}
			ImportLogger.Info("finish generate roles")
		}

		if len(resource.Functions) > 0 {
			ImportLogger.Info("start generate functions")
			captureFunc := ImportDecorateFunc(resource.Functions, func(item objects.Function, input generator.GenerateInput) bool {
				if i, ok := input.BindData.(generator.GenerateRpcData); ok {
					if i.Name == utils.SnakeCaseToPascalCase(item.Name) {
						return true
					}
				}
				return false
			}, stateChan)
			if errGenRpc := generator.GenerateRpc(projectPath, config.ProjectName, resource.Functions, resource.Tables, captureFunc); errGenRpc != nil {
				errChan <- errGenRpc
			}
			ImportLogger.Info("finish generate functions")
		}

		if len(resource.Storages) > 0 {
			ImportLogger.Info("start generate storages")
			storageInput := storages.BuildGenerateStorageInput(resource.Storages, resource.Policies)
			captureFunc := ImportDecorateFunc(storageInput, func(item *generator.GenerateStorageInput, input generator.GenerateInput) bool {
				if i, ok := input.BindData.(generator.GenerateStoragesData); ok {
					if utils.ToSnakeCase(i.Name) == utils.ToSnakeCase(item.Bucket.Name) {
						return true
					}
				}
				return false
			}, stateChan)
			if errGenStorage := generator.GenerateStorages(projectPath, config.ProjectName, storageInput, roleMap, nativeRoleMap, captureFunc); errGenStorage != nil {
				errChan <- errGenStorage
			}
			ImportLogger.Info("finish generate storages")
		}
	}()

	go func() {
		wg.Wait()
		close(stateChan)
		close(errChan)
	}()

	for {
		select {
		case rsErr := <-errChan:
			if rsErr != nil {
				return rsErr
			}
		case saveErr := <-doneListen:
			return saveErr
		}
	}
}

func updateStateOnly(importState *state.LocalState, resource *Resource, mapModelValidationTags map[string]state.ModelValidationTag) error {
	if len(resource.Tables) > 0 {
		tableInputs := tables.BuildGenerateModelInputs(resource.Tables, resource.Policies, mapModelValidationTags)
		for i := range tableInputs {
			t := tableInputs[i]
			importState.AddTable(state.TableState{
				Table:       t.Table,
				Relation:    t.Relations,
				Policies:    t.Policies,
				ModelStruct: utils.SnakeCaseToPascalCase(t.Table.Name),
				LastUpdate:  time.Now(),
			})
		}
	}

	if len(resource.Roles) > 0 {
		for i := range resource.Roles {
			r := resource.Roles[i]
			importState.AddRole(state.RoleState{
				Role:       r,
				IsNative:   false,
				RoleStruct: utils.SnakeCaseToPascalCase(r.Name),
				LastUpdate: time.Now(),
			})
		}
	}

	if len(resource.Functions) > 0 {
		for i := range resource.Functions {
			f := resource.Functions[i]
			importState.AddRpc(state.RpcState{
				Function:   f,
				RpcStruct:  utils.SnakeCaseToPascalCase(f.Name),
				LastUpdate: time.Now(),
			})
		}
	}

	if len(resource.Storages) > 0 {
		storageInputs := storages.BuildGenerateStorageInput(resource.Storages, resource.Policies)
		for i := range storageInputs {
			s := storageInputs[i]
			importState.AddStorage(state.StorageState{
				Storage:       s.Bucket,
				StorageStruct: utils.SnakeCaseToPascalCase(s.Bucket.Name),
				Policies:      s.Policies,
				LastUpdate:    time.Now(),
			})
		}
	}

	return importState.Persist()
}

func ImportDecorateFunc[T any](data []T, findFunc func(T, generator.GenerateInput) bool, stateChan chan any) generator.GenerateFn {
	return func(input generator.GenerateInput, writer io.Writer) error {
		if err := generator.Generate(input, writer); err != nil {
			return err
		}
		if rs, found := FindImportResource(data, input, findFunc); found {
			stateChan <- map[string]any{
				"item":  rs,
				"input": input,
			}
		}
		return nil
	}
}

func FindImportResource[T any](data []T, input generator.GenerateInput, findFunc func(item T, inputData generator.GenerateInput) bool) (item T, found bool) {
	for i := range data {
		t := data[i]
		if findFunc(t, input) {
			return t, true
		}
	}
	return
}

// ----- Update imported data in local state -----
func UpdateLocalStateFromImport(localState *state.LocalState, stateChan chan any) (done chan error) {
	done = make(chan error)
	go func() {
		for rs := range stateChan {
			if rs == nil {
				continue
			}

			if rsMap, isMap := rs.(map[string]any); isMap {
				item, input := rsMap["item"], rsMap["input"]
				if item == nil || input == nil {
					continue
				}

				genInput, isGenInput := input.(generator.GenerateInput)
				if !isGenInput {
					continue
				}

				switch parseItem := item.(type) {
				case *generator.GenerateModelInput:
					tableState := state.TableState{
						Table:       parseItem.Table,
						ModelPath:   genInput.OutputPath,
						ModelStruct: utils.SnakeCaseToPascalCase(parseItem.Table.Name),
						LastUpdate:  time.Now(),
						Relation:    parseItem.Relations,
						Policies:    parseItem.Policies,
					}
					localState.AddTable(tableState)
				case objects.Role:
					roleState := state.RoleState{
						Role:       parseItem,
						RolePath:   genInput.OutputPath,
						RoleStruct: utils.SnakeCaseToPascalCase(parseItem.Name),
						IsNative:   false,
						LastUpdate: time.Now(),
					}
					localState.AddRole(roleState)
				case objects.Function:
					rpcState := state.RpcState{
						Function:   parseItem,
						RpcPath:    genInput.OutputPath,
						RpcStruct:  utils.SnakeCaseToPascalCase(parseItem.Name),
						LastUpdate: time.Now(),
					}
					localState.AddRpc(rpcState)
				case *generator.GenerateStorageInput:
					storageState := state.StorageState{
						Storage:       parseItem.Bucket,
						StoragePath:   genInput.OutputPath,
						StorageStruct: utils.SnakeCaseToPascalCase(parseItem.Bucket.Name),
						Policies:      parseItem.Policies,
						LastUpdate:    time.Now(),
					}
					localState.AddStorage(storageState)
				case objects.Type:
					typeState := state.TypeState{
						Type:       parseItem,
						TypePath:   genInput.OutputPath,
						TypeStruct: utils.SnakeCaseToPascalCase(parseItem.Name),
						LastUpdate: time.Now(),
					}
					localState.AddType(typeState)
				}
			}
		}
		done <- localState.Persist()
	}()
	return done
}

// ----- Print import report -----
type ImportReport struct {
	Table    int
	Role     int
	Rpc      int
	Storage  int
	Types    int
	Policies int
}

func PrintImportReport(report ImportReport, dryRun bool) {
	var message string
	if !dryRun {
		message = "import process is complete, your code is up to date"
		if report.Role > 0 || report.Rpc > 0 || report.Storage > 0 || report.Table > 0 || report.Policies > 0 {
			message = "import process is complete, adding several new resources to the codebase"
			ImportLogger.Info(message, "Table", report.Table, "Role", report.Role, "Rpc", report.Rpc, "Storage", report.Storage, "Policies", report.Policies)
			return
		}
		ImportLogger.Info(message)
	} else {
		message = "finish running import in dry run mode, your code is up to date"
		if report.Role > 0 || report.Rpc > 0 || report.Storage > 0 || report.Table > 0 || report.Policies > 0 {
			message = "finish running import in dry run mode and add several resource"
			ImportLogger.Info(message, "Table", report.Table, "Role", report.Role, "Rpc", report.Rpc, "Storage", report.Storage, "Policies", report.Policies)
			return
		}
		ImportLogger.Info(message)
	}
}
