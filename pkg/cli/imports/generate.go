package imports

import (
	"sync"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/generator"
)

// The `generateResource` function generates various resources such as table, roles, policy and etc
// also generate framework resource like controller, route, main function and etc
func generateResource(config *raiden.Config, projectPath string, resource *Resource) error {
	if err := generator.CreateInternalFolder(projectPath); err != nil {
		return err
	}

	wg, errChan := sync.WaitGroup{}, make(chan error)

	go func() {
		wg.Wait()
		close(errChan)
	}()

	// TODO : enhance table relations
	if len(resource.Tables) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// generate all model from cloud / pg-meta
			if err := generator.GenerateModels(projectPath, resource.Tables, resource.Policies, generator.Generate); err != nil {
				errChan <- err
			} else {
				errChan <- nil
			}
		}()
	}

	// TODO : generate roles as struct
	if len(resource.Roles) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// generate all roles from cloud / pg-meta
			if err := generator.GenerateRoles(projectPath, resource.Roles, generator.Generate); err != nil {
				errChan <- err
			} else {
				errChan <- nil
			}
		}()
	}

	// TODO : generate rpc
	if len(resource.Functions) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// TODO : generate all function from cloud / pg-meta
			errChan <- nil
		}()
	}

	for rsErr := range errChan {
		if rsErr != nil {
			return rsErr
		}
	}

	return nil
}
