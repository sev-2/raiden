package acl_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/acl"
	"github.com/sev-2/raiden/pkg/mock"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

type SampleRole struct {
	raiden.RoleBase
}

func (r *SampleRole) Name() string {
	return "sample-role"
}

func signToken(t *testing.T, secret, role string) string {
	claims := jwt.MapClaims{
		"role": role,
		"exp":  time.Now().Add(time.Hour).Unix(),
		"iat":  time.Now().Add(-time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	require.NoError(t, err)
	return signed
}

func loadConfig() *raiden.Config {
	return &raiden.Config{
		DeploymentTarget:    raiden.DeploymentTargetCloud,
		ProjectId:           "test-project-cloud-id",
		SupabaseApiBasePath: "/v1",
		SupabaseApiUrl:      "http://supabase.cloud.com",
		SupabasePublicUrl:   "http://supabase.cloud.com",
		Mode:                raiden.BffMode,
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

func TestAuthenticatedInvalidHeader(t *testing.T) {
	cfg := loadConfig()
	cfg.JwtSecret = "jwt-secret"
	reqCtx := &fasthttp.RequestCtx{}
	mockCtx := &mock.MockContext{
		RequestCtx:       reqCtx,
		RequestContextFn: func() *fasthttp.RequestCtx { return reqCtx },
		ConfigFn:         func() *raiden.Config { return cfg },
	}

	err := acl.Authenticated(mockCtx)
	require.Error(t, err)
	var resp *raiden.ErrorResponse
	require.ErrorAs(t, err, &resp)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAuthenticatedInvalidToken(t *testing.T) {
	cfg := loadConfig()
	cfg.JwtSecret = "jwt-secret"
	reqCtx := &fasthttp.RequestCtx{}
	reqCtx.Request.Header.Set("Authorization", "Bearer invalid")

	mockCtx := &mock.MockContext{
		RequestCtx:       reqCtx,
		RequestContextFn: func() *fasthttp.RequestCtx { return reqCtx },
		ConfigFn:         func() *raiden.Config { return cfg },
	}

	err := acl.Authenticated(mockCtx)
	require.Error(t, err)
	var resp *raiden.ErrorResponse
	require.ErrorAs(t, err, &resp)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestGetAuthenticatedData(t *testing.T) {
	cfg := loadConfig()
	cfg.JwtSecret = "jwt-secret"
	signed := signToken(t, cfg.JwtSecret, acl.ServiceRoleName)

	reqCtx := &fasthttp.RequestCtx{}
	reqCtx.Request.Header.Set("Authorization", "Bearer "+signed)

	mockCtx := &mock.MockContext{
		RequestCtx:       reqCtx,
		RequestContextFn: func() *fasthttp.RequestCtx { return reqCtx },
		ConfigFn:         func() *raiden.Config { return cfg },
	}

	claims, err := acl.GetAuthenticatedData(mockCtx, true)
	require.NoError(t, err)
	require.Equal(t, acl.ServiceRoleName, claims.Role)
}

func TestGetAuthenticatedDataForbidden(t *testing.T) {
	cfg := loadConfig()
	cfg.JwtSecret = "jwt-secret"
	signed := signToken(t, cfg.JwtSecret, "member")

	reqCtx := &fasthttp.RequestCtx{}
	reqCtx.Request.Header.Set("Authorization", "Bearer "+signed)

	mockCtx := &mock.MockContext{
		RequestCtx:       reqCtx,
		RequestContextFn: func() *fasthttp.RequestCtx { return reqCtx },
		ConfigFn:         func() *raiden.Config { return cfg },
	}

	_, err := acl.GetAuthenticatedData(mockCtx, true)
	require.Error(t, err)
	var resp *raiden.ErrorResponse
	require.ErrorAs(t, err, &resp)
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestGetAuthenticatedDataMissingToken(t *testing.T) {
	cfg := loadConfig()
	cfg.JwtSecret = "jwt-secret"
	reqCtx := &fasthttp.RequestCtx{}
	reqCtx.Request.Header.Set("Authorization", "Bearer ")

	mockCtx := &mock.MockContext{
		RequestCtx:       reqCtx,
		RequestContextFn: func() *fasthttp.RequestCtx { return reqCtx },
		ConfigFn:         func() *raiden.Config { return cfg },
	}

	_, err := acl.GetAuthenticatedData(mockCtx, false)
	require.Error(t, err)
	var resp *raiden.ErrorResponse
	require.ErrorAs(t, err, &resp)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestGetAvailableRole(t *testing.T) {
	mockEmptyState := state.State{}
	err0 := state.Save(&mockEmptyState)
	assert.NoError(t, err0)

	emptyRoles, err1 := acl.GetAvailableRole()
	assert.Nil(t, err1)
	assert.Empty(t, emptyRoles)

	localRole := objects.Role{Name: "sample-role"}
	mockState := state.State{
		Roles: []state.RoleState{
			{
				Role: localRole,
			},
		},
	}
	err2 := state.Save(&mockState)
	assert.NoError(t, err2)

	roles, err := acl.GetAvailableRole()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(roles))
}
