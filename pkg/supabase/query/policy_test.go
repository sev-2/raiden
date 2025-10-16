package query

import (
	"testing"

	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
)

func TestBuildCreatePolicyQueryInsertSkipsUsing(t *testing.T) {
	definition := "\"name\" = 'course'"
	policy := objects.Policy{
		Name:       "new",
		Schema:     "public",
		Table:      "course",
		Action:     "PERMISSIVE",
		Roles:      []string{"Learner", "anon"},
		Command:    objects.PolicyCommandInsert,
		Definition: definition,
	}

	query := BuildCreatePolicyQuery(policy)

	assert.NotContains(t, query, "USING (")
	assert.Contains(t, query, "WITH CHECK (true)")
	assert.Contains(t, query, "GRANT INSERT ON \"public\".\"course\" TO \"Learner\"")
}

func TestBuildCreatePolicyQuerySelectKeepsUsing(t *testing.T) {
	policy := objects.Policy{
		Name:       "select_policy",
		Schema:     "public",
		Table:      "course",
		Action:     "PERMISSIVE",
		Roles:      []string{"learner"},
		Command:    objects.PolicyCommandSelect,
		Definition: "auth.uid() = user_id",
	}

	query := BuildCreatePolicyQuery(policy)

	assert.Contains(t, query, "USING ("+policy.Definition+")")
}

func TestBuildUpdatePolicyQueryInsertSkipsUsing(t *testing.T) {
	policy := objects.Policy{
		Schema:     "public",
		Table:      "course",
		Command:    objects.PolicyCommandInsert,
		Definition: "\"name\" = 'course'",
	}

	updateParam := objects.UpdatePolicyParam{
		Name: "new",
		ChangeItems: []objects.UpdatePolicyType{
			objects.UpdatePolicyDefinition,
		},
	}

	query := BuildUpdatePolicyQuery(policy, updateParam)

	assert.NotContains(t, query, "USING (")
}

func TestBuildCreatePolicyQueryDefaults(t *testing.T) {
	policy := objects.Policy{
		Name:    "empty_roles",
		Schema:  "public",
		Table:   "course",
		Command: objects.PolicyCommandSelect,
	}

	query := BuildCreatePolicyQuery(policy)

	assert.Contains(t, query, "TO PUBLIC")
	assert.NotContains(t, query, "GRANT SELECT ON \"public\".\"course\" TO")
}

func TestBuildUpdatePolicyQueryQuotesRoles(t *testing.T) {
	policy := objects.Policy{
		Schema:  "public",
		Table:   "course",
		Command: objects.PolicyCommandSelect,
		Roles:   []string{"Learner"},
	}

	updateParam := objects.UpdatePolicyParam{
		Name: "new",
		ChangeItems: []objects.UpdatePolicyType{
			objects.UpdatePolicyRoles,
		},
	}

	query := BuildUpdatePolicyQuery(policy, updateParam)
	assert.Contains(t, query, `TO "Learner"`)
}

func TestBuildDeletePolicyQuery(t *testing.T) {
	policy := objects.Policy{
		Name:    "some_policy",
		Schema:  "public",
		Table:   "course",
		Roles:   []string{"learner"},
		Command: objects.PolicyCommandInsert,
	}

	query := BuildDeletePolicyQuery(policy)
	assert.Contains(t, query, "DROP POLICY \"some_policy\" ON public.course;")
	assert.Contains(t, query, "REVOKE INSERT ON public.course FROM learner;")
}

func TestBuildUpdatePolicyQueryCoversBranches(t *testing.T) {
	check := "auth.uid() = user_id"
	policy := objects.Policy{
		Name:       "policy_new",
		Schema:     "public",
		Table:      "course",
		Command:    objects.PolicyCommandSelect,
		Definition: "auth.role() = 'admin'",
		Check:      &check,
		Roles:      []string{"admin"},
	}

	updateParam := objects.UpdatePolicyParam{
		Name: "policy_old",
		ChangeItems: []objects.UpdatePolicyType{
			objects.UpdatePolicyDefinition,
			objects.UpdatePolicyCheck,
			objects.UpdatePolicyName,
		},
	}

	query := BuildUpdatePolicyQuery(policy, updateParam)
	assert.Contains(t, query, "USING ("+policy.Definition+")")
	assert.Contains(t, query, "WITH CHECK ("+*policy.Check+")")
	assert.Contains(t, query, "RENAME TO policy_new")
}
