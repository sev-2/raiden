package roles

import (
	"reflect"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/supabase/objects"
)

type CompareDiffResult struct {
	Name           string
	SourceResource objects.Role
	TargetResource objects.Role
	DiffItems      objects.UpdateRoleParam
	IsConflict     bool
}

func Compare(source []objects.Role, target []objects.Role) error {
	diffResult, err := CompareList(source, target)
	if err != nil {
		return err
	}
	return PrintDiffResult(diffResult)
}

func CompareList(sourceRole, targetRole []objects.Role) (diffResult []CompareDiffResult, err error) {
	mapTargetRoles := make(map[int]objects.Role)
	for i := range targetRole {
		r := targetRole[i]
		mapTargetRoles[r.ID] = r
	}

	for i := range sourceRole {
		r := sourceRole[i]

		tr, isExist := mapTargetRoles[r.ID]
		if !isExist {
			continue
		}

		diffResult = append(diffResult, CompareItem(r, tr))
	}

	return
}

func CompareItem(source, target objects.Role) (diffResult CompareDiffResult) {

	var updateItem objects.UpdateRoleParam

	// assign diff result object
	diffResult.SourceResource = source
	diffResult.TargetResource = target

	if source.Name != target.Name {
		updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdateRoleName)
	}

	if source.ConnectionLimit != target.ConnectionLimit {
		updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdateConnectionLimit)
	}

	if source.CanBypassRLS != target.CanBypassRLS {
		updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdateRoleCanBypassRls)
	}

	if source.CanCreateDB != target.CanCreateDB {
		updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdateRoleCanCreateDb)
	}

	if source.CanCreateRole != target.CanCreateRole {
		updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdateRoleCanCreateRole)
	}

	if source.CanLogin != target.CanLogin {
		updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdateRoleCanLogin)
	}

	if (source.Config == nil && target.Config != nil) || (source.Config != nil && target.Config == nil) || (len(source.Config) != len(target.Config)) {
		updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdateRoleConfig)
	} else if source.Config != nil && target.Config != nil {
		for k, v := range source.Config {
			tv, exist := target.Config[k]
			if !exist {
				updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdateRoleConfig)
				break
			}

			svValue := reflect.ValueOf(v)
			tvValue := reflect.ValueOf(tv)

			if svValue.Kind() != tvValue.Kind() {
				updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdateRoleConfig)
				break
			}

			if svValue.Kind() == reflect.Ptr {
				if svValue.IsNil() && !tvValue.IsNil() {
					updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdateRoleConfig)
					break
				}

				if svValue.Interface() != tvValue.Interface() {
					updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdateRoleConfig)
					break
				}
			} else {
				if svValue.Interface() != tvValue.Interface() {
					updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdateRoleConfig)
					break
				}
			}
		}
	}

	if source.InheritRole != target.InheritRole {
		updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdateRoleInheritRole)
	}

	// Unsupported now, because need superuser role
	// if source.IsReplicationRole != target.IsReplicationRole {
	// 	updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdateRoleIsReplication)
	// }

	// if source.IsSuperuser != target.IsSuperuser {
	// 	updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdateRoleIsSuperUser)
	// }

	if (source.ValidUntil != nil && target.ValidUntil == nil) || (source.ValidUntil == nil && target.ValidUntil != nil) {
		updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdateRoleValidUntil)
	} else if source.ValidUntil != nil && target.ValidUntil != nil {
		sv := source.ValidUntil.Format(raiden.DefaultRoleValidUntilLayout)
		tv := target.ValidUntil.Format(raiden.DefaultRoleValidUntilLayout)

		if sv != tv {
			updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdateRoleValidUntil)
		}
	}

	diffResult.IsConflict = len(updateItem.ChangeItems) > 0
	diffResult.DiffItems = updateItem

	return
}
