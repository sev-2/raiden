package resource

import (
	"sync"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/postgres/roles"
	"github.com/sev-2/raiden/pkg/supabase"
	"github.com/sev-2/raiden/pkg/supabase/objects"
)

type Resource struct {
	Tables    []objects.Table
	Policies  objects.Policies
	Roles     []objects.Role
	Functions []objects.Function
}

// The Load function loads resources based on the provided flags and project ID, and returns a resource
// objects or an error.
func Load(flags *Flags, cfg *raiden.Config) (*Resource, error) {
	resource := &Resource{}
	loadChan := loadResource(cfg, flags)

	// The code block is iterating over the `loadChan` channel, which receives different types of Supabase
	// resources. It uses a switch statement to handle each type of resource accordingly.
	for result := range loadChan {
		switch rs := result.(type) {
		case []objects.Table:
			resource.Tables = rs
		case []objects.Role:
			resource.Roles = rs
		case objects.Policies:
			resource.Policies = rs
		case []objects.Function:
			resource.Functions = rs
		case error:
			return nil, rs
		}
	}

	return resource, nil
}

// The `loadResource` function loads different types of Supabase resources based on the flags provided
// and sends them to an output channel.
func loadResource(cfg *raiden.Config, flags *Flags) <-chan any {
	wg, outChan := sync.WaitGroup{}, make(chan any)

	go func() {
		wg.Wait()
		close(outChan)
	}()

	if flags.All() || flags.ModelsOnly {
		wg.Add(2)
		go loadSupabaseResource(&wg, cfg, outChan, func(cfg *raiden.Config) ([]objects.Table, error) {
			return supabase.GetTables(cfg, supabase.DefaultIncludedSchema)
		})

		go loadSupabaseResource(&wg, cfg, outChan, func(cfg *raiden.Config) ([]objects.Policy, error) {
			return supabase.GetPolicies(cfg)
		})
	}

	if flags.All() || flags.RolesOnly {
		wg.Add(1)
		go loadSupabaseResource(&wg, cfg, outChan, func(cfg *raiden.Config) ([]objects.Role, error) {
			return supabase.GetRoles(cfg)
		})
	}

	if flags.All() || flags.RpcOnly {
		wg.Add(1)
		go loadSupabaseResource(&wg, cfg, outChan, func(cfg *raiden.Config) ([]objects.Function, error) {
			return supabase.GetFunctions(cfg)
		})
	}

	return outChan
}

func loadSupabaseResource[T any](wg *sync.WaitGroup, cfg *raiden.Config, outChan chan any, callback func(cfg *raiden.Config) (T, error)) {
	defer wg.Done()

	rs, err := callback(cfg)
	if err != nil {
		outChan <- err
		return
	}
	outChan <- rs
}

func loadMapNativeRole() (map[string]raiden.Role, error) {
	mapRole := make(map[string]raiden.Role)
	for _, r := range roles.NativeRoles {
		mapRole[r.Name()] = r
	}

	return mapRole, nil
}
