package rpc

import (
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/resource/migrator"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase"
	"github.com/sev-2/raiden/pkg/supabase/objects"
)

type MigrateItem = migrator.MigrateItem[objects.Function, any]
type MigrateActionFunc = migrator.MigrateActionFunc[objects.Function, any]

var ActionFunc = MigrateActionFunc{
	CreateFunc: supabase.CreateFunction,
	UpdateFunc: func(cfg *raiden.Config, param objects.Function, items any) (err error) {
		return supabase.UpdateFunction(cfg, param)
	},
	DeleteFunc: supabase.DeleteFunction,
}

func BuildMigrateData(extractedLocalData state.ExtractRpcResult, supabaseData []objects.Function) (migrateData []MigrateItem, err error) {
	Logger.Info("start build migrate rpc data")
	if rs, err := BuildMigrateItem(supabaseData, extractedLocalData.Existing); err != nil {
		return migrateData, err
	} else {
		migrateData = append(migrateData, rs...)
	}

	// bind new table to migrated data
	Logger.Debug("filter new rpc data")
	if len(extractedLocalData.New) > 0 {
		for i := range extractedLocalData.New {
			t := extractedLocalData.New[i]
			migrateData = append(migrateData, MigrateItem{
				Type:    migrator.MigrateTypeCreate,
				NewData: t,
			})
		}
	}

	Logger.Debug("filter delete rpc data")
	if len(extractedLocalData.Delete) > 0 {
		for i := range extractedLocalData.Delete {
			t := extractedLocalData.Delete[i]
			isExist := false
			for i := range supabaseData {
				tt := supabaseData[i]
				if tt.Name == t.Name {
					isExist = true
					break
				}
			}

			if isExist {
				migrateData = append(migrateData, MigrateItem{
					Type:    migrator.MigrateTypeDelete,
					OldData: t,
				})
			}
		}
	}
	Logger.Info("finish build migrate rpc data")
	return
}

func BuildMigrateItem(supabaseData []objects.Function, localData []objects.Function) (migratedData []MigrateItem, err error) {
	Logger.Info("compare supabase and local resource for existing rpc data")
	result, e := CompareList(localData, supabaseData)
	if e != nil {
		err = e
		return
	}
	for i := range result {
		r := result[i]

		migrateType := migrator.MigrateTypeIgnore
		if r.IsConflict {
			migrateType = migrator.MigrateTypeUpdate
		}

		migratedData = append(migratedData, MigrateItem{
			Type:    migrateType,
			NewData: r.SourceResource,
			OldData: r.TargetResource,
		})
	}

	return
}

func Migrate(config *raiden.Config, rpc []MigrateItem, stateChan chan any, actions MigrateActionFunc) []error {
	return migrator.MigrateResource(config, rpc, stateChan, actions, migrator.DefaultMigrator)
}
