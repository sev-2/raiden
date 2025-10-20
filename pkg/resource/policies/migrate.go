package policies

import (
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/resource/migrator"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/utils"
)

type MigrateItem = migrator.MigrateItem[objects.Policy, objects.UpdatePolicyParam]
type MigrateActionFunc = migrator.MigrateActionFunc[objects.Policy, objects.UpdatePolicyParam]

var ActionFunc = MigrateActionFunc{
	CreateFunc: supabase.CreatePolicy,
	UpdateFunc: supabase.UpdatePolicy,
	DeleteFunc: supabase.DeletePolicy,
}

func BuildMigrateData(extractedLocalData state.ExtractPolicyResult, supabaseData []objects.Policy) (migrateData []MigrateItem, err error) {
	Logger.Info("start build policy migrate data")
	// compare and bind existing table to migrate data
	mapSpPolicies := make(map[string]objects.Policy)
	for i := range supabaseData {
		t := supabaseData[i]
		mapSpPolicies[policyKey(t)] = t
	}

	Logger.Debug("filter extracted data for update new local policy data")
	var removeExistingIndex []int
	var comparePolicies []objects.Policy
	for i := range extractedLocalData.Existing {
		p := extractedLocalData.Existing[i]
		if _, isExist := mapSpPolicies[policyKey(p)]; isExist {
			comparePolicies = append(comparePolicies, p)
		} else {
			removeExistingIndex = append(removeExistingIndex, i)
			extractedLocalData.New = append(extractedLocalData.New, p)
		}
	}
	if len(removeExistingIndex) > 0 {
		extractedLocalData.Existing = utils.RemoveByIndex(extractedLocalData.Existing, removeExistingIndex)
	}

	var removeNewIndex []int
	for i := range extractedLocalData.New {
		p := extractedLocalData.New[i]
		if _, isExist := mapSpPolicies[policyKey(p)]; isExist {
			comparePolicies = append(comparePolicies, p)
			extractedLocalData.Existing = append(extractedLocalData.Existing, p)
			removeNewIndex = append(removeNewIndex, i)
		}
	}
	if len(removeNewIndex) > 0 {
		extractedLocalData.New = utils.RemoveByIndex(extractedLocalData.New, removeNewIndex)
	}

	if rs, err := BuildMigrateItem(supabaseData, comparePolicies); err != nil {
		return migrateData, err
	} else {
		migrateData = append(migrateData, rs...)
	}

	// bind new table to migrated data
	Logger.Debug("filter new policy data")
	if len(extractedLocalData.New) > 0 {
		for i := range extractedLocalData.New {
			t := extractedLocalData.New[i]
			migrateData = append(migrateData, MigrateItem{
				Type:    migrator.MigrateTypeCreate,
				NewData: t,
			})
		}
	}

	Logger.Debug("filter delete policy data")
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
	Logger.Info("finish build policy migrate data")
	return
}

func BuildMigrateItem(supabaseData, localData []objects.Policy) (migrateData []MigrateItem, err error) {
	Logger.Info("compare supabase and local resource for existing policy data")
	result := CompareList(localData, supabaseData)
	for i := range result {
		r := result[i]
		migrateType := migrator.MigrateTypeIgnore
		if r.IsConflict {
			migrateType = migrator.MigrateTypeUpdate
		}

		migrateData = append(migrateData, MigrateItem{
			Type:           migrateType,
			NewData:        r.SourceResource,
			OldData:        r.TargetResource,
			MigrationItems: r.DiffItems,
		})
	}
	return
}

func Migrate(config *raiden.Config, policies []MigrateItem, stateChan chan any, actions MigrateActionFunc) []error {
	return migrator.MigratePolicy(config, policies, stateChan, actions)
}
