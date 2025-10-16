package acl

import (
	"slices"

	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase/objects"
)

func GetPolicy(role []string) (objects.Policies, error) {
	currentState, err := state.Load()
	if err != nil {
		return nil, err
	}
	var policies objects.Policies
	for _, table := range currentState.Tables {

		if len(role) == 0 {
			policies = append(policies, table.Policies...)
			continue
		}

		for _, p := range table.Policies {
			for _, rc := range role {
				if slices.Contains(p.Roles, rc) {
					policies = append(policies, p)
				}
			}

		}
	}

	return policies, nil
}
