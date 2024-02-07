package imports

import (
	"fmt"
	"sync"
	"time"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/generator"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase"
	"github.com/sev-2/raiden/pkg/utils"
)

type ImportState struct {
	State state.State
	Mutex sync.RWMutex
}

func (s *ImportState) AddTable(table state.TableState) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	s.State.Tables = append(s.State.Tables, table)
}

func (s *ImportState) AddRole(role state.RoleState) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	s.State.Roles = append(s.State.Roles, role)
}

func (s *ImportState) Persist() error {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()
	return state.Save(&s.State)
}

func ListenStateResource(importState *ImportState, stateChan chan any) (done chan error) {
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
					}
					importState.AddTable(tableState)
				case supabase.Role:
					roleState := state.RoleState{
						Role:       parseItem,
						RolePath:   genInput.OutputPath,
						RoleStruct: utils.SnakeCaseToPascalCase(parseItem.Name),
						IsNative:   false,
						LastUpdate: time.Now(),
					}
					importState.AddRole(roleState)
				}
			}
		}
		done <- importState.Persist()
	}()
	return done
}

func StateDecorateFunc[T any](data []T, findFunc func(T, generator.GenerateInput) bool, stateChan chan any) generator.GenerateFn {
	return func(input generator.GenerateInput) error {
		if err := generator.Generate(input); err != nil {
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

func loadAppResource(f *Flags) (tables []supabase.Table, roles []supabase.Role, err error) {
	// load app table
	latestState, err := state.Load()
	if err != nil {
		return tables, roles, err
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
		roles, err = state.ToRoles(latestState.Roles, false)
		if err != nil {
			return
		}
	}

	return
}

func createNativeRoleState(nativeRoleMap map[string]any) (roles []state.RoleState, err error) {
	for k, v := range nativeRoleMap {
		role, isRole := v.(*raiden.Role)
		if !isRole {
			return roles, fmt.Errorf("%s is not valid role", k)
		}

		roles = append(roles, state.RoleState{
			Role:       supabase.Role(*role),
			IsNative:   true,
			RoleStruct: utils.SnakeCaseToPascalCase(role.Name),
			LastUpdate: time.Now(),
		})
	}
	return
}
