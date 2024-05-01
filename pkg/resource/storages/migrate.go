package storages

import (
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/resource/migrator"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase"
	"github.com/sev-2/raiden/pkg/supabase/objects"
)

type MigrateItem = migrator.MigrateItem[objects.Bucket, objects.UpdateBucketParam]
type MigrateActionFunc = migrator.MigrateActionFunc[objects.Bucket, objects.UpdateBucketParam]

var ActionFunc = MigrateActionFunc{
	CreateFunc: supabase.CreateBucket,
	UpdateFunc: supabase.UpdateBucket,
	DeleteFunc: supabase.DeleteBucket,
}

func BuildMigrateData(extractedLocalData state.ExtractStorageResult, supabaseData []objects.Bucket) (migrateData []MigrateItem, err error) {
	Logger.Info("start build migrate storage data")
	// compare and bind existing table to migrate data
	mapSpStorages := make(map[string]bool)
	for i := range supabaseData {
		s := supabaseData[i]
		mapSpStorages[s.ID] = true
	}

	// filter existing table need compare or move to create new
	Logger.Debug("filter extracted data for update new local storage data")
	var compareStorages []objects.Bucket
	for i := range extractedLocalData.Existing {
		et := extractedLocalData.Existing[i]
		if _, isExist := mapSpStorages[et.ID]; isExist {
			compareStorages = append(compareStorages, et)
		} else {
			extractedLocalData.New = append(extractedLocalData.New, et)
		}
	}

	if rs, err := BuildMigrateItem(supabaseData, compareStorages); err != nil {
		return migrateData, err
	} else {
		migrateData = append(migrateData, rs...)
	}

	// bind new table to migrated data
	Logger.Debug("filter new storage data")
	if len(extractedLocalData.New) > 0 {
		for i := range extractedLocalData.New {
			t := extractedLocalData.New[i]
			migrateData = append(migrateData, MigrateItem{
				Type:    migrator.MigrateTypeCreate,
				NewData: t,
			})
		}
	}

	Logger.Debug("filter delete storage data")
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
	Logger.Info("finish build migrate storage data")
	return
}

func BuildMigrateItem(supabaseData []objects.Bucket, localData []objects.Bucket) (migratedData []MigrateItem, err error) {
	Logger.Info("compare supabase and local resource for existing storage data")
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

		r.DiffItems.OldData = r.TargetResource
		migratedData = append(migratedData, MigrateItem{
			Type:           migrateType,
			NewData:        r.SourceResource,
			OldData:        r.TargetResource,
			MigrationItems: r.DiffItems,
		})
	}

	return
}

func Migrate(config *raiden.Config, storages []MigrateItem, stateChan chan any, actions MigrateActionFunc) []error {
	return migrator.MigrateResource(config, storages, stateChan, actions, migrator.DefaultMigrator)
}
