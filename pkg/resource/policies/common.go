package policies

import (
	"strings"

	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/utils"
)

func CleanupAclExpression(policy *objects.Policy) {
	// cleanup check
	if policy.Check != nil {
		removedDoublePattern := utils.CleanDoubleColonPattern(*policy.Check)
		policy.Check = &removedDoublePattern
	}

	// cleanup definition
	if policy.Definition != "" {
		removeDoublePattern := utils.CleanDoubleColonPattern(policy.Definition)
		policy.Definition = removeDoublePattern
	}
}

func GetNewCountData(supabaseData []objects.Policy, localData state.ExtractPolicyResult) int {
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

func policyKey(p objects.Policy) string {
	sch := strings.ToLower(p.Schema)
	table := strings.ToLower(p.Table)
	name := strings.ToLower(p.Name)
	if sch == "" && table == "" {
		return name
	}
	return sch + "." + table + "." + name
}
