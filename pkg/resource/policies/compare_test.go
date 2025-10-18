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
			Schema:     "public",
			Table:      "table1",
			Action:     "PERMISSIVE",
			Command:    objects.PolicyCommandSelect,
			Definition: "def1",
			Check:      &check1,
			Roles:      []string{"role1", "role2"},
		},
		{
			Name:       "Policy2",
			Schema:     "public",
			Table:      "table2",
			Action:     "PERMISSIVE",
			Command:    objects.PolicyCommandSelect,
			Definition: "def2",
			Check:      &check2,
			Roles:      []string{"role1"},
		},
	}

	targetPolicies := []objects.Policy{
		{
			Name:       "Policy1",
			Schema:     "public",
			Table:      "table1",
			Action:     "PERMISSIVE",
			Command:    objects.PolicyCommandSelect,
			Definition: "def1",
			Check:      &check1,
			Roles:      []string{"role1", "role2"},
		},
		{
			Name:       "Policy2",
			Schema:     "public",
			Table:      "table2",
			Action:     "PERMISSIVE",
			Command:    objects.PolicyCommandSelect,
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
			Schema:     "public",
			Table:      "table1",
			Action:     "PERMISSIVE",
			Command:    objects.PolicyCommandSelect,
			Definition: "def1",
			Check:      &check1,
			Roles:      []string{"role1", "role2"},
		},
		TargetResource: objects.Policy{
			Name:       "Policy1",
			Schema:     "public",
			Table:      "table1",
			Action:     "PERMISSIVE",
			Command:    objects.PolicyCommandSelect,
			Definition: "def1",
			Check:      &check1,
			Roles:      []string{"role1", "role2"},
		},
		IsConflict: false,
		DiffItems: objects.UpdatePolicyParam{
			Name:        "Policy1",
			ChangeItems: []objects.UpdatePolicyType(nil),
			OldSchema:   "public",
			OldTable:    "table1",
			OldAction:   "PERMISSIVE",
			OldCommand:  objects.PolicyCommandSelect,
			OldRoles:    []string{"role1", "role2"},
		},
	}

	assert.Equal(t, expectedDiff, diffResult[0])
}

func TestCompareItem(t *testing.T) {
	check1 := "check1"
	diffCheck := "diff_check"

	sourcePolicy := objects.Policy{
		Name:       "Policy1",
		Schema:     "public",
		Table:      "table1",
		Action:     "PERMISSIVE",
		Command:    objects.PolicyCommandUpdate,
		Definition: "def1",
		Check:      &check1,
		Roles:      []string{"role1", "role2"},
	}

	targetPolicy := objects.Policy{
		Name:       "Policy1",
		Schema:     "public",
		Table:      "table1",
		Action:     "PERMISSIVE",
		Command:    objects.PolicyCommandUpdate,
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

	// Test schema/action/command differences
	sourcePolicy.Schema = "private"
	sourcePolicy.Table = "table_other"
	sourcePolicy.Action = "RESTRICTIVE"
	sourcePolicy.Command = objects.PolicyCommandInsert

	diffResult = policies.CompareItem(sourcePolicy, targetPolicy)
	assert.Contains(t, diffResult.DiffItems.ChangeItems, objects.UpdatePolicySchema)
	assert.Contains(t, diffResult.DiffItems.ChangeItems, objects.UpdatePolicyTable)
	assert.Contains(t, diffResult.DiffItems.ChangeItems, objects.UpdatePolicyAction)
	assert.Contains(t, diffResult.DiffItems.ChangeItems, objects.UpdatePolicyCommand)
}

func TestCompareItem_SkipDefinitionAndCheckByCommand(t *testing.T) {
	checkValue := "allowed"
	insertSource := objects.Policy{
		Name:       "InsertOnly",
		Schema:     "public",
		Table:      "table_insert",
		Action:     "PERMISSIVE",
		Command:    objects.PolicyCommandInsert,
		Definition: "def_should_be_ignored",
		Check:      &checkValue,
		Roles:      []string{"role1"},
	}

	insertTarget := insertSource
	insertTarget.Definition = "updated_definition"

	diffInsert := policies.CompareItem(insertSource, insertTarget)
	assert.False(t, diffInsert.IsConflict)
	assert.NotContains(t, diffInsert.DiffItems.ChangeItems, objects.UpdatePolicyDefinition)

	selectSource := objects.Policy{
		Name:       "SelectOnly",
		Schema:     "public",
		Table:      "table_select",
		Action:     "PERMISSIVE",
		Command:    objects.PolicyCommandSelect,
		Definition: "definition_in_use",
		Check:      &checkValue,
		Roles:      []string{"role1"},
	}

	diffCheckValue := "should_be_ignored"
	selectTarget := selectSource
	selectTarget.Check = &diffCheckValue

	diffSelect := policies.CompareItem(selectSource, selectTarget)
	assert.False(t, diffSelect.IsConflict)
	assert.NotContains(t, diffSelect.DiffItems.ChangeItems, objects.UpdatePolicyCheck)
}

func TestCompareItem_NormalizesPolicyClauses(t *testing.T) {
	localCheck := `"name" = 'course'`
	remoteCheck := `((name) = 'course')`

	localDefinition := `(auth.uid = owner_id)`
	remoteDefinition := `((auth.uid) = owner_id)`

	source := objects.Policy{
		Name:       "PolicyNormalized",
		Schema:     "public",
		Table:      "table_normalized",
		Action:     "PERMISSIVE",
		Command:    objects.PolicyCommandUpdate,
		Definition: localDefinition,
		Check:      &localCheck,
	}

	target := source
	target.Definition = remoteDefinition
	target.Check = &remoteCheck

	diff := policies.CompareItem(source, target)
	assert.False(t, diff.IsConflict)
	assert.NotContains(t, diff.DiffItems.ChangeItems, objects.UpdatePolicyDefinition)
	assert.NotContains(t, diff.DiffItems.ChangeItems, objects.UpdatePolicyCheck)

	castDefinition := `((("public"."table_normalized"."owner_id")::text) = (auth.uid)::text)`
	castCheck := `((("public"."table_normalized"."name")::text) = 'course'::text)`

	target.Definition = castDefinition
	target.Check = &castCheck

	diffWithCasts := policies.CompareItem(source, target)
	assert.False(t, diffWithCasts.IsConflict)
	assert.Empty(t, diffWithCasts.DiffItems.ChangeItems)
}

func TestCompareItem_IgnoresBucketClauseFormatting(t *testing.T) {
	source := objects.Policy{
		Name:       "StorageRead",
		Schema:     "storage",
		Table:      "objects",
		Action:     "PERMISSIVE",
		Command:    objects.PolicyCommandSelect,
		Definition: "((bucket_id = 'local_storage') AND true)",
	}

	formatted := objects.Policy{
		Name:       "StorageRead",
		Schema:     "storage",
		Table:      "objects",
		Action:     "PERMISSIVE",
		Command:    objects.PolicyCommandSelect,
		Definition: `("bucket_id" = 'local_storage') AND (TRUE)`,
	}

	diff := policies.CompareItem(source, formatted)
	assert.False(t, diff.IsConflict)
	assert.NotContains(t, diff.DiffItems.ChangeItems, objects.UpdatePolicyDefinition)
}

func TestCompareItem_NormalizesBooleanLiterals(t *testing.T) {
	source := objects.Policy{
		Name:       "BoolPolicy",
		Schema:     "public",
		Table:      "tbl",
		Action:     "PERMISSIVE",
		Command:    objects.PolicyCommandUpdate,
		Definition: "TRUE",
		Check:      stringPtr("TRUE"),
	}

	target := source
	target.Definition = "true"
	target.Check = stringPtr("true")

	diff := policies.CompareItem(source, target)
	assert.False(t, diff.IsConflict)
	assert.NotContains(t, diff.DiffItems.ChangeItems, objects.UpdatePolicyDefinition)
	assert.NotContains(t, diff.DiffItems.ChangeItems, objects.UpdatePolicyCheck)

	target.Definition = "false"
	target.Check = stringPtr("false")
	diffFalse := policies.CompareItem(source, target)
	assert.True(t, diffFalse.IsConflict)
	assert.Contains(t, diffFalse.DiffItems.ChangeItems, objects.UpdatePolicyDefinition)
	assert.Contains(t, diffFalse.DiffItems.ChangeItems, objects.UpdatePolicyCheck)
}

func stringPtr(v string) *string { return &v }

func TestCompareList_DuplicateNameDifferentSchema(t *testing.T) {
	check := "check"
	sourcePolicies := []objects.Policy{
		{
			Name:       "Shared",
			Schema:     "public",
			Table:      "table_a",
			Action:     "PERMISSIVE",
			Command:    objects.PolicyCommandSelect,
			Definition: "def",
			Check:      &check,
			Roles:      []string{"role1"},
		},
	}

	targetPolicies := []objects.Policy{
		{
			Name:       "Shared",
			Schema:     "admin",
			Table:      "table_b",
			Action:     "PERMISSIVE",
			Command:    objects.PolicyCommandSelect,
			Definition: "def",
			Check:      &check,
			Roles:      []string{"role1"},
		},
		{
			Name:       "Shared",
			Schema:     "public",
			Table:      "table_a",
			Action:     "PERMISSIVE",
			Command:    objects.PolicyCommandSelect,
			Definition: "def",
			Check:      &check,
			Roles:      []string{"role1"},
		},
	}

	diffResult := policies.CompareList(sourcePolicies, targetPolicies)
	assert.Equal(t, 1, len(diffResult))
	assert.Equal(t, "public", diffResult[0].TargetResource.Schema)
	assert.Equal(t, "table_a", diffResult[0].TargetResource.Table)
	assert.False(t, diffResult[0].IsConflict)
}
