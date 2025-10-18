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

	for _, p := range currentState.Tables {
		if len(p.Policies) == 0 {
			continue
		}

		for _, pp := range p.Policies {
			if len(role) == 0 {
				policies = append(policies, pp)
				continue
			}

			for _, rc := range role {
				if slices.Contains(pp.Roles, rc) {
					policies = append(policies, pp)
				}

			}
		}

	}

	return policies, nil
}
