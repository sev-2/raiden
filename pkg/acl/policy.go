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
	for _, p := range currentState.Policies {

		if len(role) == 0 {
			policies = append(policies, p.Policy)
			continue
		}

		for _, rc := range role {
			if slices.Contains(p.Policy.Roles, rc) {
				policies = append(policies, p.Policy)
			}

		}
	}

	return policies, nil
}
