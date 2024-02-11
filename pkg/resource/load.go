package resource

import (
	"context"
	"sync"

	"github.com/sev-2/raiden/pkg/supabase"
)

type Resource struct {
	Tables    []supabase.Table
	Policies  supabase.Policies
	Roles     []supabase.Role
	Functions []supabase.Function
}

// The Load function loads resources based on the provided flags and project ID, and returns a resource
// object or an error.
func Load(flags *Flags, projectId string) (*Resource, error) {
	var loadProjectId *string
	if projectId != "" {
		loadProjectId = &projectId
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resource := &Resource{}
	loadChan := loadResource(ctx, flags, loadProjectId)

	// The code block is iterating over the `loadChan` channel, which receives different types of Supabase
	// resources. It uses a switch statement to handle each type of resource accordingly.
	for result := range loadChan {
		switch rs := result.(type) {
		case []supabase.Table:
			resource.Tables = rs
		case []supabase.Role:
			resource.Roles = rs
		case supabase.Policies:
			resource.Policies = rs
		case []supabase.Function:
			resource.Functions = rs
		case error:
			return nil, rs
		}
	}

	return resource, nil
}

// The `loadResource` function loads different types of Supabase resources based on the flags provided
// and sends them to an output channel.
func loadResource(ctx context.Context, flags *Flags, projectId *string) <-chan any {
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
		go loadSupabaseResource[[]supabase.Table](&wg, projectId, outChan, func(pId *string) ([]supabase.Table, error) {
			return supabase.GetTables(ctx, pId)
		})

		go loadSupabaseResource[[]supabase.Policy](&wg, projectId, outChan, func(pId *string) ([]supabase.Policy, error) {
			return supabase.GetPolicies(ctx, pId)
		})
	}

	// This code block is checking if the `LoadAll()` flag or the `RolesOnly` flag is set in the `flags`
	// variable. If either of these flags is set, it adds 1 to the wait group (`wg.Add(1)`) and starts a
	// goroutine to load Supabase roles using the `loadSupabaseResource` function. The
	// `loadSupabaseResource` function takes a callback function `func(pId *string) ([]supabase.Role,
	// error)` as an argument, which is responsible for fetching the Supabase roles using the
	// `supabase.GetRoles` function. Once the roles are fetched, they are sent to the `outChan` channel.
	if flags.LoadAll() || flags.RolesOnly {
		wg.Add(1)
		go loadSupabaseResource[[]supabase.Role](&wg, projectId, outChan, func(pId *string) ([]supabase.Role, error) {
			return supabase.GetRoles(ctx, pId)
		})
	}

	// This code block is checking if the `LoadAll()` flag or the `RpcOnly` flag is set in the `flags`
	// variable. If either of these flags is set, it adds 1 to the wait group (`wg.Add(1)`) and starts a
	// goroutine to load Supabase functions using the `loadSupabaseResource` function.
	if flags.LoadAll() || flags.RpcOnly {
		wg.Add(1)
		go loadSupabaseResource[[]supabase.Function](&wg, projectId, outChan, func(pId *string) ([]supabase.Function, error) {
			return supabase.GetFunctions(ctx, pId)
		})
	}

	return outChan
}

// The function `loadSupabaseResource` loads a resource from Supabase asynchronously and sends the
// result or error to an output channel.
func loadSupabaseResource[T any](wg *sync.WaitGroup, projectId *string, outChan chan any, callback func(projectId *string) (T, error)) {
	defer wg.Done()

	rs, err := callback(projectId)
	if err != nil {
		outChan <- err
		return
	}
	outChan <- rs
}
