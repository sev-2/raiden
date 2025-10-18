package raiden_test

import (
	"strings"
	"testing"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/builder"
	"github.com/sev-2/raiden/pkg/supabase"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRuleBuildHandlesCommandSpecificClauses(t *testing.T) {
	selectRule := raiden.Rule("read").To(raiden.CommandSelect).
		Using(builder.Clause("owner = auth.uid()"))
	t.Run("select removes check", func(t *testing.T) {
		p, err := selectRule.Build("public", "course")
		require.NoError(t, err)
		assert.Equal(t, objects.PolicyCommandSelect, p.Command)
		assert.Equal(t, "owner = auth.uid()", p.Definition)
		assert.Nil(t, p.Check)
	})

	insertRule := raiden.Rule("create").To(raiden.CommandInsert).
		Using(builder.Clause("owner = auth.uid()")).
		Check(builder.Clause("owner = auth.uid()"))
	t.Run("insert keeps check only", func(t *testing.T) {
		p, err := insertRule.Build("public", "course")
		require.NoError(t, err)
		assert.Equal(t, objects.PolicyCommandInsert, p.Command)
		assert.Empty(t, p.Definition)
		if assert.NotNil(t, p.Check) {
			assert.Equal(t, "owner = auth.uid()", *p.Check)
		}
	})

	deleteRule := raiden.Rule("delete").To(raiden.CommandDelete).
		Check(builder.Clause("owner = auth.uid()"))
	t.Run("delete folds check into using", func(t *testing.T) {
		p, err := deleteRule.Build("public", "course")
		require.NoError(t, err)
		assert.Equal(t, objects.PolicyCommandDelete, p.Command)
		assert.Equal(t, "owner = auth.uid()", p.Definition)
		assert.Nil(t, p.Check)
	})

	updateRule := raiden.Rule("update").To(raiden.CommandUpdate)
	t.Run("update defaults to TRUE using", func(t *testing.T) {
		p, err := updateRule.Build("public", "course")
		require.NoError(t, err)
		assert.Equal(t, objects.PolicyCommandUpdate, p.Command)
		assert.Equal(t, "TRUE", p.Definition)
	})

	allRule := raiden.Rule("manage").To(raiden.CommandAll)
	t.Run("all defaults to TRUE using", func(t *testing.T) {
		p, err := allRule.Build("public", "course")
		require.NoError(t, err)
		assert.Equal(t, objects.PolicyCommandAll, p.Command)
		assert.Equal(t, "TRUE", p.Definition)
		if assert.NotNil(t, p.Check) {
			assert.Equal(t, "TRUE", *p.Check)
		}
	})
}

func TestAclBuildStoragePolicies(t *testing.T) {
	t.Run("adds bucket filter to definition and check", func(t *testing.T) {
		acl := raiden.Acl{}
		acl.Define(
			raiden.Rule("read objects").To(raiden.CommandSelect).
				Using(builder.Eq("owner", builder.AuthUID())),
			raiden.Rule("create objects").To(raiden.CommandInsert).
				Check(builder.Eq("owner", builder.AuthUID())),
		)

		policies, err := acl.BuildStoragePolicies("avatars")
		require.NoError(t, err)
		require.Len(t, policies, 2)

		var selectPolicy, insertPolicy *objects.Policy
		for i := range policies {
			policy := policies[i]
			switch policy.Command {
			case objects.PolicyCommandSelect:
				selectPolicy = &policy
			case objects.PolicyCommandInsert:
				insertPolicy = &policy
			}
		}

		if assert.NotNil(t, selectPolicy, "select policy not found") {
			assert.Equal(t, supabase.DefaultStorageSchema, selectPolicy.Schema)
			assert.Equal(t, supabase.DefaultObjectTable, selectPolicy.Table)
			assert.Contains(t, selectPolicy.Definition, `"bucket_id" = 'avatars'`)
			assert.Contains(t, selectPolicy.Definition, "auth.uid()")
			assert.Nil(t, selectPolicy.Check)
		}

		if assert.NotNil(t, insertPolicy, "insert policy not found") {
			assert.Equal(t, supabase.DefaultStorageSchema, insertPolicy.Schema)
			assert.Equal(t, supabase.DefaultObjectTable, insertPolicy.Table)
			if assert.NotNil(t, insertPolicy.Check) {
				assert.Contains(t, *insertPolicy.Check, `"bucket_id" = 'avatars'`)
				assert.Contains(t, *insertPolicy.Check, "auth.uid()")
			}
			assert.Empty(t, strings.TrimSpace(insertPolicy.Definition))
		}
	})

	t.Run("errors when bucket name empty", func(t *testing.T) {
		acl := raiden.Acl{}
		_, err := acl.BuildStoragePolicies("")
		require.Error(t, err)
	})
}
