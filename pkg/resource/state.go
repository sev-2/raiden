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
// temporary collect state data before save to state

type ResourceState struct {
	State      state.State
	NeedUpdate bool
	Mutex      sync.RWMutex
}

func (s *ResourceState) AddTable(table state.TableState) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	s.State.Tables = append(s.State.Tables, table)
	s.NeedUpdate = true
}

func (s *ResourceState) FindTable(tableId int) (index int, tableState state.TableState, found bool) {
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

func (s *ResourceState) UpdateTable(index int, state state.TableState) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	s.State.Tables[index] = state
	s.NeedUpdate = true
	return
}

func (s *ResourceState) DeleteTable(tableId int) {
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
	s.NeedUpdate = true
}

func (s *ResourceState) AddRole(role state.RoleState) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	s.State.Roles = append(s.State.Roles, role)
	s.NeedUpdate = true
}

func (s *ResourceState) FindRole(roleId int) (index int, roleState state.RoleState, found bool) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	found = false

	for i := range s.State.Roles {
		r := s.State.Roles[i]

		if r.Role.ID == roleId {
			found = true
			roleState = r
			index = i
			return
		}
	}
	return
}

func (s *ResourceState) UpdateRole(index int, state state.RoleState) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	s.State.Roles[index] = state
	s.NeedUpdate = true
	return
}

func (s *ResourceState) DeleteRole(roleId int) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	index := -1
	for i := range s.State.Roles {
		r := s.State.Roles[i]

		if r.Role.ID == roleId {
			index = i
			break
		}
	}

	if index == -1 {
		return
	}
	s.State.Roles = append(s.State.Roles[:index], s.State.Roles[index+1:]...)
	s.NeedUpdate = true
}

func (s *ResourceState) AddRpc(rpc state.RpcState) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	s.State.Rpc = append(s.State.Rpc, rpc)
	s.NeedUpdate = true
}

func (s *ResourceState) FindRpc(rpcId int) (index int, rpcState state.RpcState, found bool) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	found = false

	for i := range s.State.Rpc {
		r := s.State.Rpc[i]

		if r.Function.ID == rpcId {
			found = true
			rpcState = r
			index = i
			return
		}
	}
	return
}

func (s *ResourceState) DeleteRpc(rpcId int) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	index := -1
	for i := range s.State.Rpc {
		r := s.State.Rpc[i]

		if r.Function.ID == rpcId {
			index = i
			break
		}
	}

	if index == -1 {
		return
	}
	s.State.Rpc = append(s.State.Rpc[:index], s.State.Rpc[index+1:]...)
	s.NeedUpdate = true
}

func (s *ResourceState) Persist() error {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()
	if s.NeedUpdate {
		if err := state.Save(&s.State); err != nil {
			return err
		}
		s.NeedUpdate = false
	}
	return nil
}

func ListenImportResource(resourceState *ResourceState, stateChan chan any) (done chan error) {
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

func ListenApplyResource(projectPath string, resourceState *ResourceState, stateChan chan any) (done chan error) {
	done = make(chan error)
	go func() {
		for rs := range stateChan {
			if rs == nil {
				continue
			}

			switch m := rs.(type) {
			case *MigrateItem[objects.Table, objects.UpdateTableParam]:
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
					resourceState.UpdateTable(fIndex, tState)
				}
			case *MigrateItem[objects.Role, objects.UpdateRoleParam]:
				switch m.Type {
				case MigrateTypeCreate:
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

					logger.Debugf("add role %s to state", r.Role.Name)
					resourceState.AddRole(r)
				case MigrateTypeDelete:
					if m.OldData.Name == "" {
						continue
					}
					logger.Debugf("delete role %s from state", m.OldData.Name)
					resourceState.DeleteRole(m.OldData.ID)
				case MigrateTypeUpdate:
					fIndex, rState, found := resourceState.FindRole(m.NewData.ID)
					if !found {
						continue
					}

					logger.Debugf("update role %s in state", m.NewData.Name)
					rState.Role = m.NewData
					rState.LastUpdate = time.Now()
					resourceState.UpdateRole(fIndex, rState)
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

func loadState() (*state.State, error) {
	return state.Load()
}

func extractAppResource(f *Flags, latestState *state.State) (extractedTable state.ExtractTableResult, extractedRole state.ExtractRoleResult, rpc []objects.Function, err error) {
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
		logger.Debug("Extract role from state and app resource")
		extractedRole, err = state.ExtractRole(latestState.Roles, registeredRoles, false)
		if err != nil {
			return
		}
	}

	if f.All() || f.RpcOnly {
		rpc, err = state.ExtractRpc(latestState.Rpc, registeredRpc)
		if err != nil {
			return
		}
	}

	return
}
