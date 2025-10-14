package acl

import (
	"slices"

	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase/objects"
)

func GetPermissions(role []string) (objects.Policies, error) {
	currentState, err := state.Load()
	if err != nil {
		return nil, err
	}
	var permissions objects.Policies
	for _, table := range currentState.Tables {

		if len(role) == 0 {
			permissions = append(permissions, table.Policies...)
			continue
		}

		for _, p := range table.Policies {
			for _, rc := range role {
				if slices.Contains(p.Roles, rc) {
					permissions = append(permissions, p)
				}
			}

		}
	}

	return permissions, nil
}
