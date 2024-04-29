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
	localState, err := state.Load()
	if err != nil {
		return err
	}
	ImportLogger.Info("finish - load resource from local state")

	ImportLogger.Info("start - extract data from local state")
	appTables, appRoles, appRpcFunctions, appStorage, err := extractAppResource(flags, localState)
	if err != nil {
		return err
	}
	ImportLogger.Info("finish - extract data from local state")

	importState := state.LocalState{
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

		if err := tables.Compare(spResource.Tables, compareTables); err != nil {
			return err
		}
		ImportLogger.Debug("finish - compare table")
	}

	if (flags.All() || flags.RolesOnly) && len(appRoles.Existing) > 0 {
		ImportLogger.Debug("start - compare role")
		if err := roles.Compare(spResource.Roles, appRoles.Existing); err != nil {
			return err
		}
		ImportLogger.Debug("finish - compare role")
	}

	if (flags.All() || flags.RpcOnly) && len(appRpcFunctions.Existing) > 0 {
		ImportLogger.Debug("start - compare rpc")
		if err := rpc.Compare(spResource.Functions, appRpcFunctions.Existing); err != nil {
			return err
		}
		ImportLogger.Debug("finish - compare rpc")
	}

	if (flags.All() || flags.StoragesOnly) && len(appStorage.Existing) > 0 {
		ImportLogger.Debug("start - compare storage")
		if err := storages.Compare(spResource.Storages, appStorage.Existing); err != nil {
			return err
		}
		ImportLogger.Debug("finish - compare storage")
	}
	ImportLogger.Info("finish - compare supabase resource and local resource")

	// generate resource
	if err := generateImportResource(config, &importState, flags.ProjectPath, spResource); err != nil {
		return err
	}

	return nil
}

func generateImportResource(config *raiden.Config, importState *state.LocalState, projectPath string, resource *Resource) error {
	if err := generator.CreateInternalFolder(projectPath); err != nil {
		return err
	}

	wg, errChan, stateChan := sync.WaitGroup{}, make(chan error), make(chan any)
	doneListen := UpdateLocalStateFromImport(importState, stateChan)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if len(resource.Tables) > 0 {
			tableInputs := tables.BuildGenerateModelInputs(resource.Tables, resource.Policies)
			ImportLogger.Info("start - generate tables")
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
			ImportLogger.Info("finish - generate tables")
		}

		// generate all roles from cloud / pg-meta
		if len(resource.Roles) > 0 {
			ImportLogger.Info("start - generate roles")
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
			ImportLogger.Info("finish - generate roles")
		}

		if len(resource.Functions) > 0 {
			ImportLogger.Info("start - generate functions")
			captureFunc := ImportDecorateFunc(resource.Functions, func(item objects.Function, input generator.GenerateInput) bool {
				if i, ok := input.BindData.(generator.GenerateRpcData); ok {
					if i.Name == utils.SnakeCaseToPascalCase(item.Name) {
						return true
					}
				}
				return false
			}, stateChan)
			if errGenRpc := generator.GenerateRpc(projectPath, config.ProjectName, resource.Functions, captureFunc); errGenRpc != nil {
				errChan <- errGenRpc
			}
			ImportLogger.Info("finish - generate roles")
		}

		if len(resource.Storages) > 0 {
			ImportLogger.Info("start - generate storages")
			captureFunc := ImportDecorateFunc(resource.Storages, func(item objects.Bucket, input generator.GenerateInput) bool {
				if i, ok := input.BindData.(generator.GenerateStoragesData); ok {
					if utils.ToSnakeCase(i.Name) == utils.ToSnakeCase(item.Name) {
						return true
					}
				}
				return false
			}, stateChan)
			if errGenStorage := generator.GenerateStorages(projectPath, resource.Storages, captureFunc); errGenStorage != nil {
				errChan <- errGenStorage
			}
			ImportLogger.Info("finish - generate storages")
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
				case objects.Bucket:
					storageState := state.StorageState{
						Bucket:     parseItem,
						LastUpdate: time.Now(),
					}
					localState.AddStorage(storageState)
				}
			}
		}
		done <- localState.Persist()
	}()
	return done
}

func ImportDecorateFunc[T any](data []T, findFunc func(T, generator.GenerateInput) bool, stateChan chan any) generator.GenerateFn {
	return func(input generator.GenerateInput, writer io.Writer) error {
		if err := generator.Generate(input, nil); err != nil {
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
