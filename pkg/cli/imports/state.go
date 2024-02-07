package imports

import (
	"time"

	"github.com/sev-2/raiden/pkg/generator"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase"
	"github.com/sev-2/raiden/pkg/utils"
)

func ListenStateResource(stateChan chan any) (done chan error) {
	done = make(chan error)
	go func() {
		var newState state.State
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
					newState.Tables = append(newState.Tables, tableState)
				case supabase.Role:
					roleState := state.RoleState{
						Role:       parseItem,
						RolePath:   genInput.OutputPath,
						RoleStruct: utils.SnakeCaseToPascalCase(parseItem.Name),
						LastUpdate: time.Now(),
					}
					newState.Roles = append(newState.Roles, roleState)
				}
			}
		}
		done <- state.Save(&newState)
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
