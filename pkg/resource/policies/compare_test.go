package policies_test

import (
	"testing"

	"github.com/sev-2/raiden/pkg/resource/policies"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
)

func TestCompareList(t *testing.T) {
	check1 := "check1"
	check2 := "check2"
	diffCheck := "diff_check"

	sourcePolicies := []objects.Policy{
		{
			Name:       "Policy1",
			Definition: "def1",
			Check:      &check1,
			Roles:      []string{"role1", "role2"},
		},
		{
			Name:       "Policy2",
			Definition: "def2",
			Check:      &check2,
			Roles:      []string{"role1"},
		},
	}

	targetPolicies := []objects.Policy{
		{
			Name:       "Policy1",
			Definition: "def1",
			Check:      &check1,
			Roles:      []string{"role1", "role2"},
		},
		{
			Name:       "Policy2",
			Definition: "diff_def",
			Check:      &diffCheck,
			Roles:      []string{"role3"},
		},
	}

	diffResult := policies.CompareList(sourcePolicies, targetPolicies)
	assert.Equal(t, 2, len(diffResult))

	expectedDiff := policies.CompareDiffResult{
		Name: "",
		SourceResource: objects.Policy{
			Name:       "Policy1",
			Definition: "def1",
			Check:      &check1,
			Roles:      []string{"role1", "role2"},
		},
		TargetResource: objects.Policy{
			Name:       "Policy1",
			Definition: "def1",
			Check:      &check1,
			Roles:      []string{"role1", "role2"},
		},
		IsConflict: false,
		DiffItems: objects.UpdatePolicyParam{
			Name:        "Policy1",
			ChangeItems: []objects.UpdatePolicyType(nil),
		},
	}

	assert.Equal(t, expectedDiff, diffResult[0])
}

func TestCompareItem(t *testing.T) {
	check1 := "check1"
	diffCheck := "diff_check"

	sourcePolicy := objects.Policy{
		Name:       "Policy1",
		Definition: "def1",
		Check:      &check1,
		Roles:      []string{"role1", "role2"},
	}

	targetPolicy := objects.Policy{
		Name:       "Policy1",
		Definition: "def1",
		Check:      &check1,
		Roles:      []string{"role1", "role2"},
	}

	diffResult := policies.CompareItem(sourcePolicy, targetPolicy)
	assert.False(t, diffResult.IsConflict)
	assert.Equal(t, 0, len(diffResult.DiffItems.ChangeItems))

	// Test with differences
	targetPolicy.Definition = "diff_def"
	targetPolicy.Check = &diffCheck
	targetPolicy.Roles = []string{"role3"}

	diffResult = policies.CompareItem(sourcePolicy, targetPolicy)
	assert.True(t, diffResult.IsConflict)
	assert.Equal(t, 3, len(diffResult.DiffItems.ChangeItems))
	assert.Contains(t, diffResult.DiffItems.ChangeItems, objects.UpdatePolicyDefinition)
	assert.Contains(t, diffResult.DiffItems.ChangeItems, objects.UpdatePolicyCheck)
	assert.Contains(t, diffResult.DiffItems.ChangeItems, objects.UpdatePolicyRoles)
}
