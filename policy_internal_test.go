package raiden

import (
	"testing"

	"github.com/sev-2/raiden/pkg/builder"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/require"
)

func TestAclConfigureAndBuildPolicies(t *testing.T) {
	acl := (&Acl{}).Enable().Forced()
	require.True(t, acl.IsEnable())
	require.True(t, acl.IsForced())

	calls := 0
	acl.InitOnce(func() { calls++ })
	acl.InitOnce(func() { calls++ })
	require.Equal(t, 1, calls)

	rule := Rule("select_own").For("member").To(CommandSelect).Using(builder.Clause("owner = auth.uid()"))
	acl.Define(rule)

	policies, err := acl.BuildPolicies("public", "profiles")
	require.NoError(t, err)
	require.Len(t, policies, 1)
	require.Equal(t, objects.PolicyCommandSelect, policies[0].Command)
	require.Contains(t, policies[0].Roles, "member")
}

func TestRuleFlags(t *testing.T) {
	rule := Rule("manage").WithRestrictive().WithPermissive()
	require.Equal(t, Command(""), rule.command)

	rule.To(CommandDelete).Check(builder.True)
	rule.For("service")

	policy, err := rule.Build("", "events")
	require.NoError(t, err)
	require.Equal(t, objects.PolicyCommandDelete, policy.Command)
	require.Contains(t, policy.Roles, "service")
}

func TestEnsureHelpers(t *testing.T) {
	require.Equal(t, "TRUE", ensureClause("", true))
	require.Equal(t, "custom", ensureClause("custom", true))

	chk := ensureCheck("", false)
	require.Nil(t, chk)

	chk = ensureCheck("", true)
	require.NotNil(t, chk)
	require.Equal(t, "TRUE", *chk)
}
