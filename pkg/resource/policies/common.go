package policies

import (
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
