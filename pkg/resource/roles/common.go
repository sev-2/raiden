package roles

import (
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase/objects"
)

func GetNewCountData(supabaseData []objects.Role, localData state.ExtractRoleResult) int {
	var newCount int

	mapData := localData.ToDeleteFlatMap()
	for i := range supabaseData {
		r := supabaseData[i]

		if _, exist := mapData[r.Name]; exist {
			newCount++
		}
	}

	return newCount
}

func AttachInherithRole(mapNativeRole map[string]raiden.Role, supabaseData objects.Roles, RoleMembership objects.RoleMemberships) []objects.Role {
	mapRoles, mapRoleMemberships := supabaseData.ToMap(), RoleMembership.GroupByInheritId()
	for i, r := range supabaseData {
		// find membership key
		inheritRoles, exist := mapRoleMemberships[r.ID]
		if !exist {
			continue
		}

		// find available role
		inheritCandidate := []*objects.Role{}
		for _, ih := range inheritRoles {
			if _, isExist := mapNativeRole[ih.InheritRole]; isExist {
				continue
			}

			if rc, exist := mapRoles[ih.ParentID]; exist {
				if _, isExist := mapNativeRole[rc.Name]; isExist {
					continue
				}

				inheritCandidate = append(inheritCandidate, &rc)
			}
		}

		// assign inherith role
		r.InheritRoles = inheritCandidate
		supabaseData[i] = r
	}
	return supabaseData
}
