package acl_test

import (
	"testing"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/acl"
	"github.com/sev-2/raiden/pkg/mock"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
)

type SampleRole struct {
	raiden.RoleBase
}

func (r *SampleRole) Name() string {
	return "sample-role"
}

func loadConfig() *raiden.Config {
	return &raiden.Config{
		DeploymentTarget:    raiden.DeploymentTargetCloud,
		ProjectId:           "test-project-cloud-id",
		SupabaseApiBasePath: "/v1",
		SupabaseApiUrl:      "http://supabase.cloud.com",
		SupabasePublicUrl:   "http://supabase.cloud.com",
	}
}

func TestSetUserRole(t *testing.T) {
	config := loadConfig()
	roleBase := SampleRole{}
	someRole := &roleBase
	err := acl.SetUserRole(config, "user-id", someRole)
	assert.Error(t, err)

	mockSupabase := &mock.MockSupabase{Cfg: config}
	mockSupabase.Activate()
	defer mockSupabase.Deactivate()

	someUser := objects.User{ID: "sample-id", Role: "authenticated"}
	err0 := mockSupabase.MockAdminUpdateUserDataWithExpectedResponse(200, someUser)
	assert.NoError(t, err0)

	localRole := objects.Role{Name: "sample-role"}
	mockState := state.State{
		Roles: []state.RoleState{
			{
				Role: localRole,
			},
		},
	}
	err1 := state.Save(&mockState)
	assert.NoError(t, err1)

	err2 := acl.SetUserRole(config, "user-id", someRole)
	assert.NoError(t, err2)
}

func TestGetAvailableRole(t *testing.T) {
	localRole := objects.Role{Name: "sample-role"}
	mockState := state.State{
		Roles: []state.RoleState{
			{
				Role: localRole,
			},
		},
	}
	err1 := state.Save(&mockState)
	assert.NoError(t, err1)

	roles, err := acl.GetAvailableRole()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(roles))
}
