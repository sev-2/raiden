package state

import (
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/supabase/objects"
)

type ExtractPolicyResult struct {
	Existing []objects.Policy
	New      []objects.Policy
	Delete   []objects.Policy
}

func ExtractPolicy(policyStates []PolicyState, appPolicies []raiden.Policy) (result ExtractPolicyResult, err error) {
	mapPolicyState := map[string]PolicyState{}
	for i := range policyStates {
		r := policyStates[i]
		mapPolicyState[r.Policy.Name] = r
	}

	for _, policy := range appPolicies {
		state, isStateExist := mapPolicyState[policy.Name()]
		p, err := raiden.BuildPolicy(policy)
		if err != nil {
			return result, err
		}
		if !isStateExist {
			result.New = append(result.New, *p)
			continue
		}

		pr := BuildPolicyFromState(state, p)
		result.Existing = append(result.Existing, pr)

		delete(mapPolicyState, policy.Name())
	}

	for _, state := range mapPolicyState {
		result.Delete = append(result.Delete, state.Policy)
	}
	return
}

func BuildPolicyFromState(ps PolicyState, policy *objects.Policy) (p objects.Policy) {
	p = ps.Policy
	p.Table = policy.Table
	p.Name = policy.Name
	p.Action = policy.Action
	p.Check = policy.Check
	p.Command = policy.Command
	p.Definition = policy.Definition
	p.Roles = policy.Roles
	p.Schema = policy.Schema

	return
}

func (ep ExtractPolicyResult) ToDeleteFlatMap() map[string]*objects.Policy {
	mapData := make(map[string]*objects.Policy)

	if len(ep.Delete) > 0 {
		for i := range ep.Delete {
			r := ep.Delete[i]
			mapData[r.Name] = &r
		}
	}

	return mapData
}
