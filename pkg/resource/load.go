package resource

import (
	"sync"

	"github.com/hashicorp/go-hclog"
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/connector/pgmeta"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/postgres/roles"
	"github.com/sev-2/raiden/pkg/resource/policies"
	"github.com/sev-2/raiden/pkg/supabase"
	"github.com/sev-2/raiden/pkg/supabase/objects"
)

var LoadLogger hclog.Logger = logger.HcLog().Named("import.load")

type Resource struct {
	Tables          []objects.Table
	Policies        objects.Policies
	Roles           []objects.Role
	RoleMemberships []objects.RoleMembership
	Functions       []objects.Function
	Storages        []objects.Bucket
	Indexes         []objects.Index
	RelationActions []objects.TablesRelationshipAction
	Types           []objects.Type
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
			LoadLogger.Debug("finish get Table from server")
		case []objects.Role:
			resource.Roles = rs
			LoadLogger.Debug("finish get Role from server")
		case []objects.RoleMembership:
			resource.RoleMemberships = rs
			LoadLogger.Debug("finish get Role Membership from server")
		case objects.Policies:
			resource.Policies = rs
			LoadLogger.Debug("finish get Policy from server")
		case []objects.Function:
			resource.Functions = rs
			LoadLogger.Debug("finish get Function from server")
		case []objects.Bucket:
			resource.Storages = rs
			LoadLogger.Debug("finish get Bucket from server")
		case []objects.Index:
			resource.Indexes = rs
			LoadLogger.Debug("finish get Indexes from server")
		case []objects.TablesRelationshipAction:
			resource.RelationActions = rs
			LoadLogger.Debug("finish get Relation Action from server")
		case []objects.Type:
			resource.Types = rs
			LoadLogger.Debug("finish get Type from server")
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

	if cfg.Mode == raiden.BffMode {
		loadBffResources(&wg, flags, cfg, outChan)
	} else {
		loadServiceResources(&wg, flags, cfg, outChan)
	}

	return outChan
}

func loadBffResources(wg *sync.WaitGroup, flags *Flags, cfg *raiden.Config, outChan chan any) {
	if flags.All() || flags.ModelsOnly || flags.StoragesOnly {
		wg.Add(1)
		LoadLogger.Debug("get Policy from server")
		go loadDatabaseResource(wg, cfg, outChan, func(cfg *raiden.Config) (objects.Policies, error) {
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
		LoadLogger.Debug("get Role from server")
		go loadDatabaseResource(wg, cfg, outChan, func(cfg *raiden.Config) ([]objects.Role, error) {
			return supabase.GetRoles(cfg)
		})
	}

	if (flags.All() || flags.ModelsOnly) && cfg.AllowedTables != "" {
		wg.Add(1)
		LoadLogger.Debug("get Types from server")
		go loadDatabaseResource(wg, cfg, outChan, func(cfg *raiden.Config) ([]objects.Type, error) {
			return supabase.GetTypes(cfg, []string{raiden.DefaultTypeSchema})
		})

		wg.Add(1)
		LoadLogger.Debug("get Table from server")
		go loadDatabaseResource(wg, cfg, outChan, func(cfg *raiden.Config) ([]objects.Table, error) {
			return supabase.GetTables(cfg, supabase.DefaultIncludedSchema)
		})

		wg.Add(1)
		LoadLogger.Debug("get Index from server")
		go loadDatabaseResource(wg, cfg, outChan, func(cfg *raiden.Config) ([]objects.Index, error) {
			return supabase.GetIndexes(cfg, supabase.DefaultIncludedSchema[0])
		})

		wg.Add(1)
		LoadLogger.Debug("get Table Relation Actions from server")
		go loadDatabaseResource(wg, cfg, outChan, func(cfg *raiden.Config) ([]objects.TablesRelationshipAction, error) {
			return supabase.GetTableRelationshipActions(cfg, supabase.DefaultIncludedSchema[0])
		})
	}

	if flags.All() || flags.RolesOnly {
		wg.Add(1)
		LoadLogger.Debug("get Role from server")
		go loadDatabaseResource(wg, cfg, outChan, func(cfg *raiden.Config) ([]objects.Role, error) {
			return supabase.GetRoles(cfg)
		})

		wg.Add(1)
		LoadLogger.Debug("get Role Membership from server")
		go loadDatabaseResource(wg, cfg, outChan, func(cfg *raiden.Config) ([]objects.RoleMembership, error) {
			return supabase.GetRoleMemberships(cfg, supabase.DefaultIncludedSchema)
		})
	}

	if flags.All() || flags.RpcOnly {
		if flags.RpcOnly {
			wg.Add(1)
			LoadLogger.Debug("get Types from server")
			go loadDatabaseResource(wg, cfg, outChan, func(cfg *raiden.Config) ([]objects.Type, error) {
				return supabase.GetTypes(cfg, []string{raiden.DefaultTypeSchema})
			})

			wg.Add(1)
			LoadLogger.Debug("get Table from server")
			go loadDatabaseResource(wg, cfg, outChan, func(cfg *raiden.Config) ([]objects.Table, error) {
				return supabase.GetTables(cfg, supabase.DefaultIncludedSchema)
			})
		}

		wg.Add(1)
		LoadLogger.Debug("get Function from server")
		go loadDatabaseResource(wg, cfg, outChan, func(cfg *raiden.Config) ([]objects.Function, error) {
			return supabase.GetFunctions(cfg)
		})
	}

	if flags.All() || flags.StoragesOnly {
		wg.Add(1)
		LoadLogger.Debug("get Bucket from server")
		go loadDatabaseResource(wg, cfg, outChan, func(cfg *raiden.Config) ([]objects.Bucket, error) {
			return supabase.GetBuckets(cfg)
		})
	}
}

func loadServiceResources(wg *sync.WaitGroup, flags *Flags, cfg *raiden.Config, outChan chan any) {
	if flags.All() || flags.RolesOnly {
		wg.Add(1)
		LoadLogger.Debug("Get Role From Pg Meta")
		go loadDatabaseResource(wg, cfg, outChan, func(cfg *raiden.Config) ([]objects.Role, error) {
			return supabase.GetRoles(cfg)
		})

		wg.Add(1)
		LoadLogger.Debug("Get Role Membership From Pg Meta")
		go loadDatabaseResource(wg, cfg, outChan, func(cfg *raiden.Config) ([]objects.RoleMembership, error) {
			return supabase.GetRoleMemberships(cfg, supabase.DefaultIncludedSchema)
		})
	}

	wg.Add(1)
	LoadLogger.Debug("Get Type From Pg Meta")
	go loadDatabaseResource(wg, cfg, outChan, func(cfg *raiden.Config) ([]objects.Type, error) {
		return pgmeta.GetTypes(cfg, []string{raiden.DefaultTypeSchema})
	})

	wg.Add(1)
	LoadLogger.Debug("Get Table From Pg Meta")
	go loadDatabaseResource(wg, cfg, outChan, func(cfg *raiden.Config) ([]objects.Table, error) {
		return pgmeta.GetTables(cfg, []string{"public"}, true)
	})

	wg.Add(1)
	LoadLogger.Debug("Get Index From Pg Meta")
	go loadDatabaseResource(wg, cfg, outChan, func(cfg *raiden.Config) ([]objects.Index, error) {
		return pgmeta.GetIndexes(cfg, "public")
	})

	wg.Add(1)
	LoadLogger.Debug("Get Table Relation Actions From Pg Meta")
	go loadDatabaseResource(wg, cfg, outChan, func(cfg *raiden.Config) ([]objects.TablesRelationshipAction, error) {
		return pgmeta.GetTableRelationshipActions(cfg, "public")
	})

	wg.Add(1)
	LoadLogger.Debug("Get Function From Pg Meta")
	go loadDatabaseResource(wg, cfg, outChan, func(cfg *raiden.Config) ([]objects.Function, error) {
		return pgmeta.GetFunctions(cfg)
	})
}

func loadDatabaseResource[T any](wg *sync.WaitGroup, cfg *raiden.Config, outChan chan any, callback func(cfg *raiden.Config) (T, error)) {
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
