package resource

import (
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
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/utils"
)

var ImportLogger hclog.Logger = logger.HcLog().Named("import")

// List of import resource
// [x] import table, relation, column specification and acl
// [x] import role
// [x] import function
// [x] import storage
func Import(flags *Flags, config *raiden.Config) error {
	if flags.DryRun {
		ImportLogger.Info("running import in dry run mode")
	}

	// load map native role
	ImportLogger.Info("load native role")
	mapNativeRole, err := loadMapNativeRole()
	if err != nil {
		return err
	}

	// load supabase resource
	ImportLogger.Info("load resource from supabase")
	spResource, err := Load(flags, config)
	if err != nil {
		return err
	}
	spResource.Tables = tables.AttachIndexAndAction(spResource.Tables, spResource.Indexes, spResource.RelationActions)

	// create import state
	ImportLogger.Debug("get native roles")
	nativeStateRoles := filterIsNativeRole(mapNativeRole, spResource.Roles)

	// filter table for with allowed schema
	ImportLogger.Debug("start filter table and function by allowed schema", "allowed-schema", flags.AllowedSchema)
	ImportLogger.Trace("filter table by schema")
	spResource.Tables = filterTableBySchema(spResource.Tables, strings.Split(flags.AllowedSchema, ",")...)

	ImportLogger.Trace("filter function by schema")
	spResource.Functions = filterFunctionBySchema(spResource.Functions, strings.Split(flags.AllowedSchema, ",")...)
	ImportLogger.Debug("finish filter table and function by allowed schema")

	ImportLogger.Trace("remove native role for supabase list role")
	spResource.Roles = filterUserRole(spResource.Roles, mapNativeRole)

	// load app resource
	ImportLogger.Info("load resource from local state")
	localState, err := state.Load()
	if err != nil {
		return err
	}

	ImportLogger.Info("extract data from local state")
	appTables, appRoles, appRpcFunctions, appStorage, err := extractAppResource(flags, localState)
	if err != nil {
		return err
	}

	importState := state.LocalState{
		State: state.State{
			Roles: nativeStateRoles,
		},
	}

	// dry run import errors
	dryRunError := []string{}
	mapModelValidationTags := make(map[string]state.ModelValidationTag)

	// compare resource
	ImportLogger.Info("compare supabase resource and local resource")
	if (flags.All() || flags.ModelsOnly) && len(appTables.Existing) > 0 {
		if !flags.DryRun {
			ImportLogger.Debug("start compare table")
		}

		for i := range appTables.New {
			nt := appTables.New[i]
			if nt.ValidationTags != nil {
				mapModelValidationTags[nt.Table.Name] = nt.ValidationTags
			}
		}

		// compare table
		var compareTables []objects.Table
		for i := range appTables.Existing {
			et := appTables.Existing[i]
			compareTables = append(compareTables, et.Table)

			if et.ValidationTags != nil {
				mapModelValidationTags[et.Table.Name] = et.ValidationTags
			}
		}

		if err := tables.Compare(spResource.Tables, compareTables); err != nil {
			if flags.DryRun {
				dryRunError = append(dryRunError, err.Error())
			} else {
				return err
			}
		}
		if !flags.DryRun {
			ImportLogger.Debug("finish compare table")
		}
	}

	if (flags.All() || flags.RolesOnly) && len(appRoles.Existing) > 0 {
		if !flags.DryRun {
			ImportLogger.Debug("start compare role")
		}
		if err := roles.Compare(spResource.Roles, appRoles.Existing); err != nil {
			if flags.DryRun {
				dryRunError = append(dryRunError, err.Error())
			} else {
				return err
			}
		}
		if !flags.DryRun {
			ImportLogger.Debug("finish compare role")
		}
	}

	if (flags.All() || flags.RpcOnly) && len(appRpcFunctions.Existing) > 0 {
		if !flags.DryRun {
			ImportLogger.Debug("start compare rpc")
		}
		if err := rpc.Compare(spResource.Functions, appRpcFunctions.Existing); err != nil {
			if flags.DryRun {
				dryRunError = append(dryRunError, err.Error())
			} else {
				return err
			}
		}
		if !flags.DryRun {
			ImportLogger.Debug("finish compare rpc")
		}
	}

	if (flags.All() || flags.StoragesOnly) && len(appStorage.Existing) > 0 {
		if !flags.DryRun {
			ImportLogger.Debug("start compare storage")
		}

		// compare table
		var compareStorages []objects.Bucket
		for i := range appStorage.Existing {
			es := appStorage.Existing[i]
			compareStorages = append(compareStorages, es.Storage)
		}

		if err := storages.Compare(spResource.Storages, compareStorages); err != nil {
			if flags.DryRun {
				dryRunError = append(dryRunError, err.Error())
			} else {
				return err
			}
		}
		if !flags.DryRun {
			ImportLogger.Debug("finish compare storage")
		}
	}

	// import report
	importReport := ImportReport{
		Role:    roles.GetNewCountData(spResource.Roles, appRoles),
		Table:   tables.GetNewCountData(spResource.Tables, appTables),
		Storage: storages.GetNewCountData(spResource.Storages, appStorage),
		Rpc:     rpc.GetNewCountData(spResource.Functions, appRpcFunctions),
	}
	if !flags.DryRun {
		if flags.UpdateStateOnly {
			return updateStateOnly(&importState, spResource, mapModelValidationTags)
		} else {
			// generate resource
			if err := generateImportResource(config, &importState, flags.ProjectPath, spResource, mapModelValidationTags); err != nil {
				return err
			}
			PrintImportReport(importReport, false)
		}

	} else {

		if len(dryRunError) > 0 {
			errMessage := strings.Join(dryRunError, "\n")
			ImportLogger.Error("got error", "err-msg", errMessage)
			return nil
		}
		PrintImportReport(importReport, true)
	}

	return nil
}

// ----- Generate import data -----
func generateImportResource(config *raiden.Config, importState *state.LocalState, projectPath string, resource *Resource, mapModelValidationTags map[string]state.ModelValidationTag) error {
	if err := generator.CreateInternalFolder(projectPath); err != nil {
		return err
	}

	wg, errChan, stateChan := sync.WaitGroup{}, make(chan error), make(chan any)
	doneListen := UpdateLocalStateFromImport(importState, stateChan)

	wg.Add(1)
	go func() {
		defer wg.Done()
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

			if err := generator.GenerateModels(projectPath, tableInputs, captureFunc); err != nil {
				errChan <- err
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
			if errGenStorage := generator.GenerateStorages(projectPath, storageInput, captureFunc); errGenStorage != nil {
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
				}
			}
		}
		done <- localState.Persist()
	}()
	return done
}

// ----- Print import report -----
type ImportReport struct {
	Table   int
	Role    int
	Rpc     int
	Storage int
}

func PrintImportReport(report ImportReport, dryRun bool) {
	var message string
	if !dryRun {
		message = "import process is complete, your code is up to date"
		if report.Role > 0 || report.Rpc > 0 || report.Storage > 0 || report.Table > 0 {
			message = "import process is complete, adding several new resources to the codebase"
			ImportLogger.Info(message, "Table", report.Table, "Role", report.Role, "Rpc", report.Rpc, "Storage", report.Storage)
			return
		}
		ImportLogger.Info(message)
	} else {
		message = "finish running import in dry run mode, your code is up to date"
		if report.Role > 0 || report.Rpc > 0 || report.Storage > 0 || report.Table > 0 {
			message = "finish running import in dry run mode and add several resource"
			ImportLogger.Info(message, "Table", report.Table, "Role", report.Role, "Rpc", report.Rpc, "Storage", report.Storage)
			return
		}
		ImportLogger.Info(message)
	}
}
