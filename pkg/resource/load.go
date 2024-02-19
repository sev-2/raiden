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

	// The code block is checking if the `LoadAll()` flag or the `ModelsOnly` flag is set in the `flags`
	// variable. If either of these flags is set, it adds 2 to the wait group (`wg.Add(2)`) and starts two
	// goroutines to load Supabase resources.
	if flags.LoadAll() || flags.ModelsOnly {
		wg.Add(2)
		go loadSupabaseResource[[]objects.Table](&wg, cfg, outChan, func(cfg *raiden.Config) ([]objects.Table, error) {
			return supabase.GetTables(cfg, supabase.DefaultIncludedSchema)
		})

		go loadSupabaseResource[[]objects.Policy](&wg, cfg, outChan, func(cfg *raiden.Config) ([]objects.Policy, error) {
			return supabase.GetPolicies(cfg)
		})
	}

	// This code block is checking if the `LoadAll()` flag or the `RolesOnly` flag is set in the `flags`
	// variable. If either of these flags is set, it adds 1 to the wait group (`wg.Add(1)`) and starts a
	// goroutine to load Supabase roles using the `loadSupabaseResource` function. The
	// `loadSupabaseResource` function takes a callback function `func(pId *string) ([]objects.Role,
	// error)` as an argument, which is responsible for fetching the Supabase roles using the
	// `objects.GetRoles` function. Once the roles are fetched, they are sent to the `outChan` channel.
	if flags.LoadAll() || flags.RolesOnly {
		wg.Add(1)
		go loadSupabaseResource[[]objects.Role](&wg, cfg, outChan, func(cfg *raiden.Config) ([]objects.Role, error) {
			return supabase.GetRoles(cfg)
		})
	}

	// This code block is checking if the `LoadAll()` flag or the `RpcOnly` flag is set in the `flags`
	// variable. If either of these flags is set, it adds 1 to the wait group (`wg.Add(1)`) and starts a
	// goroutine to load Supabase functions using the `loadSupabaseResource` function.
	if flags.LoadAll() || flags.RpcOnly {
		wg.Add(1)
		go loadSupabaseResource[[]objects.Function](&wg, cfg, outChan, func(cfg *raiden.Config) ([]objects.Function, error) {
			return supabase.GetFunctions(cfg)
		})
	}

	return outChan
}

// The function `loadSupabaseResource` loads a resource from Supabase asynchronously and sends the
// result or error to an output channel.
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
