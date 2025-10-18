package policies_test

import (
	"testing"

	"github.com/sev-2/raiden/pkg/resource/policies"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
)

func TestCleanupAclExpression(t *testing.T) {
	t.Run("Cleanup both Check and Definition", func(t *testing.T) {
		check := "some_check_expression::text"
		definition := "some_definition_expression::text"
		policy := &objects.Policy{
			Check:      &check,
			Definition: definition,
		}

		policies.CleanupAclExpression(policy)

		assert.Equal(t, "some_check_expression", *policy.Check)
		assert.Equal(t, "some_definition_expression", policy.Definition)
	})

	t.Run("Cleanup Check only", func(t *testing.T) {
		check := "another_check_expression::int"
		policy := &objects.Policy{
			Check: &check,
		}

		policies.CleanupAclExpression(policy)

		assert.Equal(t, "another_check_expression", *policy.Check)
		assert.Empty(t, policy.Definition)
	})

	t.Run("Cleanup Definition only", func(t *testing.T) {
		definition := "another_definition_expression::int"
		policy := &objects.Policy{
			Definition: definition,
		}

		policies.CleanupAclExpression(policy)

		assert.Nil(t, policy.Check)
		assert.Equal(t, "another_definition_expression", policy.Definition)
	})

	t.Run("No Cleanup needed", func(t *testing.T) {
		check := "check_expression"
		definition := "definition_expression"
		policy := &objects.Policy{
			Check:      &check,
			Definition: definition,
		}

		policies.CleanupAclExpression(policy)

		assert.Equal(t, "check_expression", *policy.Check)
		assert.Equal(t, "definition_expression", policy.Definition)
	})

	t.Run("Nil Check and Definition", func(t *testing.T) {
		policy := &objects.Policy{}

		policies.CleanupAclExpression(policy)

		assert.Nil(t, policy.Check)
		assert.Equal(t, "", policy.Definition)
	})
}

func TestGetNewCountData(t *testing.T) {
    supabasePolicies := []objects.Policy{{Name: "p1"}, {Name: "p2"}, {Name: "p3"}}
    local := state.ExtractPolicyResult{
        Delete: []objects.Policy{{Name: "p1"}, {Name: "p3"}},
    }

    count := policies.GetNewCountData(supabasePolicies, local)
    assert.Equal(t, 2, count)

    count = policies.GetNewCountData(nil, state.ExtractPolicyResult{})
    assert.Equal(t, 0, count)
}
