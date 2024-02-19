package resource

import (
	"io"
	"sync"
	"time"

	"github.com/sev-2/raiden/pkg/generator"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/utils"
)

type resourceState struct {
	State state.State
	Mutex sync.RWMutex
}

func (s *resourceState) AddTable(table state.TableState) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	s.State.Tables = append(s.State.Tables, table)
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

func ListenStateResource(resourceState *resourceState, stateChan chan any) (done chan error) {
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

func StateDecorateFunc[T any](data []T, findFunc func(T, generator.GenerateInput) bool, stateChan chan any) generator.GenerateFn {
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

func loadAppResource(f *Flags) (tables []objects.Table, roles []objects.Role, rpc []objects.Function, err error) {
	// load app table
	latestState, err := state.Load()
	if err != nil {
		return tables, roles, rpc, err
	}

	if latestState == nil {
		return
	}

	if f.LoadAll() || f.ModelsOnly {
		tables, err = state.ToTables(latestState.Tables)
		if err != nil {
			return
		}
	}

	if f.LoadAll() || f.RolesOnly {
		roles, err = state.ToRoles(latestState.Roles, registeredRoles, false)
		if err != nil {
			return
		}
	}

	if f.LoadAll() || f.RpcOnly {
		rpc, err = state.ToRpc(latestState.RpcState, registeredRpc)
		if err != nil {
			return
		}
	}

	return
}
