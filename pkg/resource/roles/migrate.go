package roles

import (
	"strings"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/resource/migrator"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase"
	"github.com/sev-2/raiden/pkg/supabase/objects"
)

type MigrateItem = migrator.MigrateItem[objects.Role, objects.UpdateRoleParam]
type MigrateActionFunc = migrator.MigrateActionFunc[objects.Role, objects.UpdateRoleParam]

var ActionFunc = MigrateActionFunc{
	CreateFunc: supabase.CreateRole, UpdateFunc: supabase.UpdateRole, DeleteFunc: supabase.DeleteRole,
}

func BuildMigrateData(extractedLocalData state.ExtractRoleResult, supabaseData []objects.Role) (migrateData []MigrateItem, err error) {
	Logger.Info("start build migrate role data")
	// compare and bind existing table to migrate data
	mapSpRole := make(map[int]bool)
	for i := range supabaseData {
		t := supabaseData[i]
		mapSpRole[t.ID] = true
	}

	// filter existing table need compare or move to create new
	Logger.Debug("filter extracted data for update new local role data")
	var compareRoles []objects.Role
	for i := range extractedLocalData.Existing {
		et := extractedLocalData.Existing[i]
		if _, isExist := mapSpRole[et.ID]; isExist {
			compareRoles = append(compareRoles, et)
		} else {
			extractedLocalData.New = append(extractedLocalData.New, et)
		}
	}

	if rs, err := BuildMigrateItem(supabaseData, compareRoles); err != nil {
		return migrateData, err
	} else {
		migrateData = append(migrateData, rs...)
	}

	// bind new table to migrated data
	Logger.Debug("filter new role data")
	if len(extractedLocalData.New) > 0 {
		for i := range extractedLocalData.New {
			t := extractedLocalData.New[i]
			migrateItem := MigrateItem{
				Type:    migrator.MigrateTypeCreate,
				NewData: t,
			}

			inheritItems := buildGrantInheritItems(t)
			if len(inheritItems) > 0 {
				migrateItem.MigrationItems.ChangeInheritItems = inheritItems
			}

			migrateData = append(migrateData, migrateItem)
		}
	}

	Logger.Debug("filter delete role data")
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
	Logger.Info("finish build migrate role data")
	return
}

func BuildMigrateItem(supabaseData []objects.Role, localData []objects.Role) (migrateData []MigrateItem, err error) {
	Logger.Info("compare supabase and local resource for existing role data")
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
		migrateData = append(migrateData, MigrateItem{
			Type:           migrateType,
			NewData:        r.SourceResource,
			OldData:        r.TargetResource,
			MigrationItems: r.DiffItems,
		})
	}

	return
}

func Migrate(config *raiden.Config, roles []MigrateItem, stateChan chan any, actions MigrateActionFunc) []error {
	return migrator.MigrateResource(config, roles, stateChan, actions, migrateRole)
}

func migrateRole(params migrator.MigrateFuncParam[objects.Role, objects.UpdateRoleParam]) error {
	switch params.Data.Type {
	case migrator.MigrateTypeCreate:
		return migrateCreateRole(params)
	case migrator.MigrateTypeUpdate:
		if err := params.ActionFuncs.UpdateFunc(params.Config, params.Data.NewData, params.Data.MigrationItems); err != nil {
			return err
		}

		params.StateChan <- &params.Data
		return nil
	case migrator.MigrateTypeDelete:
		if err := params.ActionFuncs.DeleteFunc(params.Config, params.Data.OldData); err != nil {
			return err
		}

		params.StateChan <- &params.Data
		return nil
	default:
		return nil
	}
}

func migrateCreateRole(params migrator.MigrateFuncParam[objects.Role, objects.UpdateRoleParam]) error {
	desired := params.Data.NewData
	created, err := params.ActionFuncs.CreateFunc(params.Config, desired)
	if err != nil {
		return err
	}

	// make sure inherit metadata is reflected in created record
	created.InheritRole = desired.InheritRole
	created.InheritRoles = cloneInheritRoles(desired.InheritRoles)

	inheritItems := params.Data.MigrationItems.ChangeInheritItems
	if len(inheritItems) == 0 {
		inheritItems = buildGrantInheritItems(desired)
	}

	if len(inheritItems) > 0 {
		updateParam := objects.UpdateRoleParam{
			OldData:            created,
			ChangeInheritItems: inheritItems,
		}
		if err := params.ActionFuncs.UpdateFunc(params.Config, created, updateParam); err != nil {
			return err
		}
	}

	params.Data.NewData = created
	params.StateChan <- &params.Data
	return nil
}

func buildGrantInheritItems(role objects.Role) []objects.UpdateRoleInheritItem {
	if len(role.InheritRoles) == 0 {
		return nil
	}

	items := make([]objects.UpdateRoleInheritItem, 0, len(role.InheritRoles))
	unique := make(map[string]struct{})

	for i := range role.InheritRoles {
		inherit := role.InheritRoles[i]
		if inherit == nil {
			continue
		}

		name := strings.TrimSpace(inherit.Name)
		if name == "" {
			continue
		}

		key := strings.ToLower(name)
		if _, exist := unique[key]; exist {
			continue
		}
		unique[key] = struct{}{}

		items = append(items, objects.UpdateRoleInheritItem{
			Role: objects.Role{Name: name},
			Type: objects.UpdateRoleInheritGrant,
		})
	}

	return items
}

func cloneInheritRoles(inherits []*objects.Role) []*objects.Role {
	if len(inherits) == 0 {
		return nil
	}

	cloned := make([]*objects.Role, 0, len(inherits))
	unique := make(map[string]struct{})

	for i := range inherits {
		r := inherits[i]
		if r == nil {
			continue
		}

		name := strings.TrimSpace(r.Name)
		if name == "" {
			continue
		}

		key := strings.ToLower(name)
		if _, exist := unique[key]; exist {
			continue
		}
		unique[key] = struct{}{}

		cloned = append(cloned, &objects.Role{Name: name})
	}

	if len(cloned) == 0 {
		return nil
	}

	return cloned
}
