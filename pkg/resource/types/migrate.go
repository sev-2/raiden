package types

import (
	"fmt"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/connector/pgmeta"
	"github.com/sev-2/raiden/pkg/resource/migrator"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase"
	"github.com/sev-2/raiden/pkg/supabase/objects"
)

type MigrateItem = migrator.MigrateItem[objects.Type, objects.UpdateTypeParam]
type MigrateActionFunc = migrator.MigrateActionFunc[objects.Type, objects.UpdateTypeParam]

var ActionFunc = MigrateActionFunc{
	CreateFunc: func(cfg *raiden.Config, param objects.Type) (response objects.Type, err error) {
		if cfg.Mode == raiden.SvcMode {
			return pgmeta.CreateType(cfg, param)
		}
		return supabase.CreateType(cfg, param)
	},
	UpdateFunc: func(cfg *raiden.Config, param objects.Type, items objects.UpdateTypeParam) (err error) {
		if cfg.Mode == raiden.SvcMode {
			return pgmeta.UpdateType(cfg, param)
		}
		return supabase.UpdateType(cfg, param)
	},
	DeleteFunc: func(cfg *raiden.Config, param objects.Type) (err error) {
		if cfg.Mode == raiden.SvcMode {
			return pgmeta.DeleteType(cfg, param)
		}
		return supabase.DeleteType(cfg, param)
	},
}

func BuildMigrateData(extractedLocalData state.ExtractTypeResult, supabaseData []objects.Type) (migrateData []MigrateItem, err error) {
	Logger.Info("start build migrate type data")
	if rs, err := BuildMigrateItem(supabaseData, extractedLocalData.Existing); err != nil {
		return migrateData, err
	} else {
		migrateData = append(migrateData, rs...)
	}

	// bind new table to migrated data
	Logger.Debug("filter new type data")
	if len(extractedLocalData.New) > 0 {
		for i := range extractedLocalData.New {
			t := extractedLocalData.New[i]
			migrateData = append(migrateData, MigrateItem{
				Type:    migrator.MigrateTypeCreate,
				NewData: t,
			})
		}
	}

	Logger.Debug("filter delete type data")
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

	fmt.Println("total : ", len(migrateData))
	Logger.Info("finish build migrate type data")
	return
}

func BuildMigrateItem(supabaseData []objects.Type, localData []objects.Type) (migratedData []MigrateItem, err error) {
	Logger.Info("compare supabase and local resource for existing type data")
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

func Migrate(config *raiden.Config, dataType []MigrateItem, stateChan chan any, actions MigrateActionFunc) []error {
	return migrator.MigrateResource(config, dataType, stateChan, actions, migrator.DefaultMigrator)
}
