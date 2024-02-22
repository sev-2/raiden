package resource

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/sev-2/raiden/pkg/generator"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/utils"
)

// ---- Resource state -----
// temporary collect state date before save to state

type resourceState struct {
	State state.State
	Mutex sync.RWMutex
}

func (s *resourceState) AddTable(table state.TableState) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	s.State.Tables = append(s.State.Tables, table)
}

func (s *resourceState) FindTable(tableId int) (index int, tableState state.TableState, found bool) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	found = false

	for i := range s.State.Tables {
		t := s.State.Tables[i]

		if t.Table.ID == tableId {
			found = true
			tableState = t
			index = i
			return
		}
	}
	return
}

func (s *resourceState) DeleteTable(tableId int) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	index := -1
	for i := range s.State.Tables {
		t := s.State.Tables[i]

		if t.Table.ID == tableId {
			index = i
			break
		}
	}

	if index == -1 {
		return
	}
	s.State.Tables = append(s.State.Tables[:index], s.State.Tables[index+1:]...)
}

func (s *resourceState) AddRole(role state.RoleState) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	s.State.Roles = append(s.State.Roles, role)
}

func (s *resourceState) AddRpc(rpc state.RpcState) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	s.State.RpcState = append(s.State.RpcState, rpc)
}

func (s *resourceState) Persist() error {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()
	return state.Save(&s.State)
}

func ListenImportResource(resourceState *resourceState, stateChan chan any) (done chan error) {
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
					}
					resourceState.AddTable(tableState)
				case objects.Role:
					roleState := state.RoleState{
						Role:       parseItem,
						RolePath:   genInput.OutputPath,
						RoleStruct: utils.SnakeCaseToPascalCase(parseItem.Name),
						IsNative:   false,
						LastUpdate: time.Now(),
					}
					resourceState.AddRole(roleState)
				case objects.Function:
					rpcState := state.RpcState{
						Function:   parseItem,
						RpcPath:    genInput.OutputPath,
						RpcStruct:  utils.SnakeCaseToPascalCase(parseItem.Name),
						LastUpdate: time.Now(),
					}

					resourceState.AddRpc(rpcState)
				}
			}
		}
		done <- resourceState.Persist()
	}()
	return done
}

func ListenApplyResource(projectPath string, resourceState *resourceState, stateChan chan any) (done chan error) {
	done = make(chan error)
	go func() {
		for rs := range stateChan {
			if rs == nil {
				continue
			}
			switch m := rs.(type) {
			case *MigrateItem[objects.Table, objects.UpdateTableItem]:
				switch m.Type {
				case MigrateTypeCreate:
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

					logger.Debugf("add table %s to state", ts.Table.Name)
					resourceState.AddTable(ts)
				case MigrateTypeDelete:
					if m.OldData.Name == "" {
						continue
					}
					logger.Debugf("delete table %s from state", m.OldData.Name)
					resourceState.DeleteTable(m.OldData.ID)
				case MigrateTypeUpdate:
					fIndex, tState, found := resourceState.FindTable(m.NewData.ID)
					if !found {
						continue
					}

					logger.Debugf("update table %s in state", m.NewData.Name)
					tState.Table = m.NewData
					tState.LastUpdate = time.Now()
					resourceState.State.Tables[fIndex] = tState

				}
			}
		}
		done <- resourceState.Persist()
	}()
	return done
}

func ImportDecorateFunc[T any](data []T, findFunc func(T, generator.GenerateInput) bool, stateChan chan any) generator.GenerateFn {
	return func(input generator.GenerateInput, writer io.Writer) error {
		if err := generator.Generate(input, nil); err != nil {
			return err
		}
		if rs, found := FindResource(data, input, findFunc); found {
			stateChan <- map[string]any{
				"item":  rs,
				"input": input,
			}
		}
		return nil
	}
}

func FindResource[T any](data []T, input generator.GenerateInput, findFunc func(item T, inputData generator.GenerateInput) bool) (item T, found bool) {
	for i := range data {
		t := data[i]
		if findFunc(t, input) {
			return t, true
		}
	}
	return
}

func loadAppResource() (*state.State, error) {
	return state.Load()
}

func extractAppResourceState(f *Flags, latestState *state.State) (extractedTable state.ExtractTableResult, roles []objects.Role, rpc []objects.Function, err error) {
	if latestState == nil {
		return
	}

	if f.All() || f.ModelsOnly {
		logger.Debug("Extract table from state and app resource")
		extractedTable, err = state.ExtractTable(latestState.Tables, registeredModels)
		if err != nil {
			return
		}
	}

	if f.All() || f.RolesOnly {
		roles, err = state.ToRoles(latestState.Roles, registeredRoles, false)
		if err != nil {
			return
		}
	}

	if f.All() || f.RpcOnly {
		rpc, err = state.ToRpc(latestState.RpcState, registeredRpc)
		if err != nil {
			return
		}
	}

	return
}
