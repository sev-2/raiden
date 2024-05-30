package resource

import (
	"sync"

	"github.com/hashicorp/go-hclog"
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/postgres/roles"
	"github.com/sev-2/raiden/pkg/resource/policies"
	"github.com/sev-2/raiden/pkg/supabase"
	"github.com/sev-2/raiden/pkg/supabase/objects"
)

var LoadLogger hclog.Logger = logger.HcLog().Named("import.load")

type Resource struct {
	Tables    []objects.Table
	Policies  objects.Policies
	Roles     []objects.Role
	Functions []objects.Function
	Storages  []objects.Bucket
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
			LoadLogger.Debug("Finish Get Table From Supabase")
		case []objects.Role:
			resource.Roles = rs
			LoadLogger.Debug("Finish Get Role From Supabase")
		case objects.Policies:
			resource.Policies = rs
			LoadLogger.Debug("Finish Get Policy From Supabase")
		case []objects.Function:
			resource.Functions = rs
			LoadLogger.Debug("Finish Get Function From Supabase")
		case []objects.Bucket:
			resource.Storages = rs
			LoadLogger.Debug("Finish Get Bucket From Supabase")
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

	if flags.All() || flags.ModelsOnly || flags.StoragesOnly {
		wg.Add(1)
		LoadLogger.Debug("Get Policy From Supabase")
		go loadSupabaseResource(&wg, cfg, outChan, func(cfg *raiden.Config) (objects.Policies, error) {
			rs, e := supabase.GetPolicies(cfg)
			if e != nil {
				return rs, e
			}

			// cleanup policy expression
			var cleanedPolicies objects.Policies
			for i := range rs {
				p := rs[i]
				policies.CleanupAclExpression(&p)
				cleanedPolicies = append(cleanedPolicies, p)
			}

			return cleanedPolicies, nil
		})

		wg.Add(1)
		LoadLogger.Debug("Get Role From Supabase")
		go loadSupabaseResource(&wg, cfg, outChan, func(cfg *raiden.Config) ([]objects.Role, error) {
			return supabase.GetRoles(cfg)
		})
	}

	if flags.All() || flags.ModelsOnly {
		wg.Add(1)
		LoadLogger.Debug("Get Table From Supabase")
		go loadSupabaseResource(&wg, cfg, outChan, func(cfg *raiden.Config) ([]objects.Table, error) {
			return supabase.GetTables(cfg, supabase.DefaultIncludedSchema)
		})

	}

	if flags.All() || flags.RolesOnly {
		wg.Add(1)
		LoadLogger.Debug("Get Role From Supabase")
		go loadSupabaseResource(&wg, cfg, outChan, func(cfg *raiden.Config) ([]objects.Role, error) {
			return supabase.GetRoles(cfg)
		})
	}

	if flags.All() || flags.RpcOnly {
		wg.Add(1)
		LoadLogger.Debug("Get Function From Supabase")
		go loadSupabaseResource(&wg, cfg, outChan, func(cfg *raiden.Config) ([]objects.Function, error) {
			return supabase.GetFunctions(cfg)
		})
	}

	if flags.All() || flags.StoragesOnly {
		wg.Add(1)
		LoadLogger.Debug("Get Bucket From Supabase")
		go loadSupabaseResource(&wg, cfg, outChan, func(cfg *raiden.Config) ([]objects.Bucket, error) {
			return supabase.GetBuckets(cfg)
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
