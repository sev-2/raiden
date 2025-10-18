package acl

import (
	"testing"

	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/require"
)

func TestGetPolicy(t *testing.T) {
	st := &state.State{
		Tables: []state.TableState{
			{
				Policies: []objects.Policy{
					{
						ID:         1,
						Table:      "Orders",
						Name:       "AllowMembers",
						Command:    objects.PolicyCommandSelect,
						Roles:      []string{"member"},
						Schema:     "public",
						Action:     "SELECT",
						TableID:    1,
						Definition: "auth.role() = 'member'",
					},
					{
						ID:         2,
						Table:      "Orders",
						Name:       "AllowManagers",
						Command:    objects.PolicyCommandUpdate,
						Roles:      []string{"manager"},
						Schema:     "public",
						Action:     "UPDATE",
						TableID:    1,
						Definition: "auth.role() = 'manager'",
					},
				},
			},
			{
				Policies: []objects.Policy{},
			},
			{
				Policies: []objects.Policy{
					{
						ID:         3,
						Table:      "Payments",
						Name:       "AllowAll",
						Command:    objects.PolicyCommandAll,
						Roles:      []string{},
						Schema:     "public",
						Action:     "ALL",
						TableID:    2,
						Definition: "true",
					},
				},
			},
		},
	}

	prepareState(t, st)

	got, err := GetPolicy(nil)
	require.NoError(t, err)
	require.Len(t, got, 3)

	filtered, err := GetPolicy([]string{"member", "manager"})
	require.NoError(t, err)
	require.Len(t, filtered, 2)

	filteredNames := []string{filtered[0].Name, filtered[1].Name}
	require.ElementsMatch(t, []string{"AllowMembers", "AllowManagers"}, filteredNames)
}
