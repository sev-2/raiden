package tables

import (
	"fmt"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/resource/migrator"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase"
	"github.com/sev-2/raiden/pkg/supabase/objects"
)

type MigrateItem = migrator.MigrateItem[objects.Table, objects.UpdateTableParam]
type MigrateActionFunc = migrator.MigrateActionFunc[objects.Table, objects.UpdateTableParam]

var ActionFunc = MigrateActionFunc{
	CreateFunc: supabase.CreateTable, UpdateFunc: supabase.UpdateTable,
	DeleteFunc: func(cfg *raiden.Config, param objects.Table) (err error) {
		return supabase.DeleteTable(cfg, param, true)
	},
}

func BuildMigrateData(extractedLocalData state.ExtractTableResult, supabaseData []objects.Table, allowedTable []string) (migrateData []MigrateItem, err error) {
	Logger.Info("start build migrate table data")
	// compare and bind existing table to migrate data
	mapSpTable := make(map[int]bool)
	for i := range supabaseData {
		t := supabaseData[i]
		mapSpTable[t.ID] = true
	}

	mapAllowedTable := make(map[string]bool)
	for _, t := range allowedTable {
		if t == "" {
			continue
		}
		mapAllowedTable[t] = true
	}

	// filter existing table need compare or move to create new
	Logger.Debug("filter extracted data for update new local table data")
	var compareTables []objects.Table
	for i := range extractedLocalData.Existing {
		et := extractedLocalData.Existing[i]

		if !isAllowedTable(mapAllowedTable, et.Table.Name) {
			return migrateData, fmt.Errorf("table %s is not allowed to modify", et.Table.Name)
		}

		if _, isExist := mapSpTable[et.Table.ID]; isExist {
			compareTables = append(compareTables, et.Table)
		} else {
			extractedLocalData.New = append(extractedLocalData.New, et)
		}

	}
	if rs, err := BuildMigrateItem(supabaseData, compareTables); err != nil {
		return migrateData, err
	} else {
		migrateData = append(migrateData, rs...)
	}

	// bind new table to migrated data
	Logger.Debug("filter new table data")
	if len(extractedLocalData.New) > 0 {
		for i := range extractedLocalData.New {
			t := extractedLocalData.New[i]
			if !isAllowedTable(mapAllowedTable, t.Table.Name) {
				return migrateData, fmt.Errorf("table %s is not allowed to modify", t.Table.Name)
			}
			migrateData = append(migrateData, MigrateItem{
				Type:    migrator.MigrateTypeCreate,
				NewData: t.Table,
			})
		}
	}

	// bind delete table to migrate data
	Logger.Debug("filter delete table data")
	if len(extractedLocalData.Delete) > 0 {
		for i := range extractedLocalData.Delete {
			t := extractedLocalData.Delete[i]

			if !isAllowedTable(mapAllowedTable, t.Table.Name) {
				return migrateData, fmt.Errorf("table %s is not allowed to modify", t.Table.Name)
			}

			isExist := false
			for i := range supabaseData {
				tt := supabaseData[i]
				if tt.Name == t.Table.Name {
					isExist = true
					break
				}
			}

			if isExist {
				migrateData = append(migrateData, MigrateItem{
					Type:    migrator.MigrateTypeDelete,
					OldData: t.Table,
				})
			}
		}
	}
	Logger.Info("finish build migrate table data")
	return
}

func BuildMigrateItem(supabaseData, localData []objects.Table) (migratedData []MigrateItem, err error) {
	Logger.Info("compare supabase and local resource for existing table data")
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

func Migrate(config *raiden.Config, tables []MigrateItem, stateChan chan any, actions MigrateActionFunc) []error {
	return migrator.MigrateResource(config, tables, stateChan, actions, migrator.DefaultMigrator)
}

func isAllowedTable(mapAllowedTable map[string]bool, table string) bool {
	if len(mapAllowedTable) == 0 {
		return true
	}

	if _, exist := mapAllowedTable[table]; !exist {
		return false
	}

	return true
}
