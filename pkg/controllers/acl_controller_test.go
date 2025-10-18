package controllers_test

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/acl"
	"github.com/sev-2/raiden/pkg/controllers"
	"github.com/sev-2/raiden/pkg/mock"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/utils"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

func prepareState(t *testing.T, st *state.State) {
	t.Helper()

	relativeDir := filepath.Join("testdata", "tmp", t.Name())

	origDir := state.StateFileDir
	origName := state.StateFileName

	state.StateFileDir = relativeDir
	state.StateFileName = fmt.Sprintf("state_%s", t.Name())

	curDir, err := utils.GetCurrentDirectory()
	require.NoError(t, err)

	statePath := filepath.Join(curDir, relativeDir)
	require.NoError(t, os.MkdirAll(statePath, 0o755))

	t.Cleanup(func() {
		state.StateFileDir = origDir
		state.StateFileName = origName
		require.NoError(t, os.RemoveAll(statePath))
	})

	require.NoError(t, state.Save(st))
}

func issueToken(t *testing.T, role, secret string) string {
	t.Helper()

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

func TestAclUserRoleControllerPatchSuccess(t *testing.T) {
	cfg := loadConfig()
	cfg.JwtSecret = "jwt-secret"

	viewer := objects.Role{Name: "viewer"}
	role := objects.Role{Name: "editor", InheritRoles: []*objects.Role{&viewer}}

	prepareState(t, &state.State{
		Roles: []state.RoleState{{Role: role}, {Role: viewer}},
	})

	mockSupabase := &mock.MockSupabase{Cfg: cfg}
	mockSupabase.Activate()
	t.Cleanup(mockSupabase.Deactivate)

	require.NoError(t, mockSupabase.MockAdminUpdateUserDataWithExpectedResponse(200, objects.User{ID: "user-id", Role: "editor"}))

	token := issueToken(t, acl.ServiceRoleName, cfg.JwtSecret)
	reqCtx := &fasthttp.RequestCtx{}
	reqCtx.Request.Header.Set("Authorization", "Bearer "+token)

	var response controllers.AclUserRoleResponse
	var sendErrorCalled bool

	ctx := &mock.MockContext{
		RequestCtx: reqCtx,
		ConfigFn: func() *raiden.Config {
			return cfg
		},
		SendJsonFn: func(data any) error {
			response = data.(controllers.AclUserRoleResponse)
			return nil
		},
		SendErrorFn: func(message string) error {
			sendErrorCalled = true
			return nil
		},
		SendErrorWithCodeFn: func(statusCode int, err error) error {
			sendErrorCalled = true
			return nil
		},
		RequestContextFn: func() *fasthttp.RequestCtx {
			return reqCtx
		},
		Data: make(map[string]any),
	}

	controller := controllers.AclUserRoleController{
		Payload: &controllers.AclUserRoleRequest{
			UserUuid: "user-id",
			Role:     "editor",
		},
	}

	err := controller.Patch(ctx)
	require.NoError(t, err)
	require.False(t, sendErrorCalled)
	require.Equal(t, "success update role user to editor", response.Message)
}

func TestAclUserRoleControllerPatchUnknownRole(t *testing.T) {
	cfg := loadConfig()
	cfg.JwtSecret = "jwt-secret"

	prepareState(t, &state.State{
		Roles: []state.RoleState{{Role: objects.Role{Name: "editor"}}},
	})

	token := issueToken(t, acl.ServiceRoleName, cfg.JwtSecret)
	reqCtx := &fasthttp.RequestCtx{}
	reqCtx.Request.Header.Set("Authorization", "Bearer "+token)

	var sendErrorCalled bool
	var sendErrorMsg string

	ctx := &mock.MockContext{
		RequestCtx: reqCtx,
		ConfigFn: func() *raiden.Config {
			return cfg
		},
		SendJsonFn: func(data any) error {
			return nil
		},
		SendErrorFn: func(message string) error {
			sendErrorCalled = true
			sendErrorMsg = message
			return nil
		},
		SendErrorWithCodeFn: func(statusCode int, err error) error {
			sendErrorCalled = true
			return nil
		},
		RequestContextFn: func() *fasthttp.RequestCtx {
			return reqCtx
		},
		Data: make(map[string]any),
	}

	controller := controllers.AclUserRoleController{
		Payload: &controllers.AclUserRoleRequest{
			UserUuid: "user-456",
			Role:     "ghost",
		},
	}

	err := controller.Patch(ctx)
	require.NoError(t, err)
	require.True(t, sendErrorCalled)
	require.Equal(t, "role ghost is not available", sendErrorMsg)
}

func TestAclUserRoleControllerPatchUnauthorized(t *testing.T) {
	cfg := loadConfig()
	cfg.JwtSecret = "jwt-secret"
	reqCtx := &fasthttp.RequestCtx{}
	mockCtx := &mock.MockContext{
		RequestCtx:       reqCtx,
		RequestContextFn: func() *fasthttp.RequestCtx { return reqCtx },
		ConfigFn:         func() *raiden.Config { return cfg },
	}

	controller := controllers.AclUserRoleController{Payload: &controllers.AclUserRoleRequest{UserUuid: "user-id", Role: "editor"}}

	err := controller.Patch(mockCtx)
	require.Error(t, err)
	var resp *raiden.ErrorResponse
	require.ErrorAs(t, err, &resp)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAclUserRoleControllerPatchSetRoleFailure(t *testing.T) {
	cfg := loadConfig()
	cfg.JwtSecret = "jwt-secret"

	viewer := objects.Role{Name: "viewer"}
	role := objects.Role{Name: "editor", InheritRoles: []*objects.Role{&viewer}}

	prepareState(t, &state.State{
		Roles: []state.RoleState{{Role: role}, {Role: viewer}},
	})

	mockSupabase := &mock.MockSupabase{Cfg: cfg}
	mockSupabase.Activate()
	t.Cleanup(mockSupabase.Deactivate)

	require.NoError(t, mockSupabase.MockAdminUpdateUserDataWithExpectedResponse(500, objects.User{}))

	token := issueToken(t, acl.ServiceRoleName, cfg.JwtSecret)
	reqCtx := &fasthttp.RequestCtx{}
	reqCtx.Request.Header.Set("Authorization", "Bearer "+token)

	var sendErrorCalled bool
	var sendErrorMsg string

	mockCtx := &mock.MockContext{
		RequestCtx:       reqCtx,
		RequestContextFn: func() *fasthttp.RequestCtx { return reqCtx },
		ConfigFn:         func() *raiden.Config { return cfg },
		SendJsonFn:       func(data any) error { return nil },
		SendErrorFn: func(message string) error {
			sendErrorCalled = true
			sendErrorMsg = message
			return nil
		},
		SendErrorWithCodeFn: func(statusCode int, err error) error {
			sendErrorCalled = true
			return nil
		},
		Data: make(map[string]any),
	}

	controller := controllers.AclUserRoleController{
		Payload: &controllers.AclUserRoleRequest{
			UserUuid: "user-id",
			Role:     "editor",
		},
	}

	err := controller.Patch(mockCtx)
	require.NoError(t, err)
	require.True(t, sendErrorCalled)
	require.Equal(t, "failed to User role", sendErrorMsg)
}

func TestAclRoleControllerGetSuccess(t *testing.T) {
	cfg := loadConfig()
	cfg.JwtSecret = "jwt-secret"

	child := objects.Role{Name: "viewer"}
	parent := objects.Role{Name: "editor", InheritRoles: []*objects.Role{&child}}

	prepareState(t, &state.State{
		Roles: []state.RoleState{{Role: parent}, {Role: child}},
	})

	token := issueToken(t, acl.ServiceRoleName, cfg.JwtSecret)
	reqCtx := &fasthttp.RequestCtx{}
	reqCtx.Request.Header.Set("Authorization", "Bearer "+token)

	var response controllers.AclRoleResponse

	ctx := &mock.MockContext{
		RequestCtx: reqCtx,
		ConfigFn:   func() *raiden.Config { return cfg },
		SendJsonFn: func(data any) error {
			response = data.(controllers.AclRoleResponse)
			return nil
		},
		SendErrorFn:         func(message string) error { return nil },
		SendErrorWithCodeFn: func(statusCode int, err error) error { return err },
		RequestContextFn:    func() *fasthttp.RequestCtx { return reqCtx },
		Data:                make(map[string]any),
	}

	controller := controllers.AclRoleController{Payload: &controllers.AclRoleRequest{}}

	err := controller.Get(ctx)
	require.NoError(t, err)
	require.Len(t, response, 2)
	require.Equal(t, "editor", response[0].Name)
	require.Len(t, response[0].InheritRoles, 1)
	require.Equal(t, "viewer", response[0].InheritRoles[0].Name)
	require.Equal(t, "viewer", response[1].Name)
	require.Empty(t, response[1].InheritRoles)
}

func TestAclRoleControllerGetUnauthorized(t *testing.T) {
	cfg := loadConfig()
	cfg.JwtSecret = "jwt-secret"
	reqCtx := &fasthttp.RequestCtx{}
	mockCtx := &mock.MockContext{
		RequestCtx:       reqCtx,
		RequestContextFn: func() *fasthttp.RequestCtx { return reqCtx },
		ConfigFn:         func() *raiden.Config { return cfg },
	}

	controller := controllers.AclRoleController{Payload: &controllers.AclRoleRequest{}}

	err := controller.Get(mockCtx)
	require.Error(t, err)
	var resp *raiden.ErrorResponse
	require.ErrorAs(t, err, &resp)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAclUserPolicyControllerGetForMember(t *testing.T) {
	cfg := loadConfig()
	cfg.JwtSecret = "jwt-secret"

	viewer := objects.Role{Name: "viewer"}
	member := objects.Role{Name: "member", InheritRoles: []*objects.Role{&viewer}}

	prepareState(t, &state.State{
		Roles: []state.RoleState{{Role: member}, {Role: viewer}},
		Tables: []state.TableState{
			{
				Table: objects.Table{Name: "Orders"},
				Policies: []objects.Policy{
					{Table: "Orders", Name: "OrdersRead", Command: objects.PolicyCommandSelect, Roles: []string{"member"}},
					{Table: "Orders", Name: "OrdersUpdate", Command: objects.PolicyCommandUpdate, Roles: []string{"manager"}},
				},
			},
		},
	})

	token := issueToken(t, "member", cfg.JwtSecret)
	reqCtx := &fasthttp.RequestCtx{}
	reqCtx.Request.Header.Set("Authorization", "Bearer "+token)

	var response controllers.AclUserPolicyResponse
	var sendErrorCalled bool

	ctx := &mock.MockContext{
		RequestCtx: reqCtx,
		ConfigFn:   func() *raiden.Config { return cfg },
		SendJsonFn: func(data any) error {
			response = data.(controllers.AclUserPolicyResponse)
			return nil
		},
		SendErrorFn: func(message string) error {
			sendErrorCalled = true
			return nil
		},
		SendErrorWithCodeFn: func(statusCode int, err error) error {
			sendErrorCalled = true
			return nil
		},
		RequestContextFn: func() *fasthttp.RequestCtx { return reqCtx },
		Data:             make(map[string]any),
	}

	controller := controllers.AclUserPolicyController{Payload: &controllers.AclUserPolicyRequest{}}

	err := controller.Get(ctx)
	require.NoError(t, err)
	require.False(t, sendErrorCalled)
	require.Len(t, response, 1)
	require.Contains(t, response, "orders.select.orders_read")
}

func TestAclUserPolicyControllerGetForbiddenWhenRoleMissing(t *testing.T) {
	cfg := loadConfig()
	cfg.JwtSecret = "jwt-secret"

	prepareState(t, &state.State{
		Roles: []state.RoleState{{Role: objects.Role{Name: "member"}}},
	})

	token := issueToken(t, "unknown", cfg.JwtSecret)
	reqCtx := &fasthttp.RequestCtx{}
	reqCtx.Request.Header.Set("Authorization", "Bearer "+token)

	var sendErrorCalled bool
	var statusCode int

	ctx := &mock.MockContext{
		RequestCtx: reqCtx,
		ConfigFn:   func() *raiden.Config { return cfg },
		SendJsonFn: func(data any) error { return nil },
		SendErrorFn: func(message string) error {
			sendErrorCalled = true
			return nil
		},
		SendErrorWithCodeFn: func(code int, err error) error {
			sendErrorCalled = true
			statusCode = code
			return nil
		},
		RequestContextFn: func() *fasthttp.RequestCtx { return reqCtx },
		Data:             make(map[string]any),
	}

	controller := controllers.AclUserPolicyController{Payload: &controllers.AclUserPolicyRequest{}}

	err := controller.Get(ctx)
	require.NoError(t, err)
	require.True(t, sendErrorCalled)
	require.Equal(t, fasthttp.StatusForbidden, statusCode)
}

func TestAclUserPolicyControllerGetServiceRole(t *testing.T) {
	cfg := loadConfig()
	cfg.JwtSecret = "jwt-secret"

	policyState := &state.State{
		Roles: []state.RoleState{{Role: objects.Role{Name: "editor"}}},
		Tables: []state.TableState{
			{
				Table: objects.Table{Name: "Orders"},
				Policies: []objects.Policy{
					{Table: "Orders", Name: "OrdersRead", Command: objects.PolicyCommandSelect, Roles: []string{"editor"}},
					{Table: "Orders", Name: "OrdersWrite", Command: objects.PolicyCommandUpdate, Roles: []string{"editor"}},
				},
			},
		},
	}

	prepareState(t, policyState)

	token := issueToken(t, acl.ServiceRoleName, cfg.JwtSecret)
	reqCtx := &fasthttp.RequestCtx{}
	reqCtx.Request.Header.Set("Authorization", "Bearer "+token)

	var response controllers.AclUserPolicyResponse

	mockCtx := &mock.MockContext{
		RequestCtx:       reqCtx,
		RequestContextFn: func() *fasthttp.RequestCtx { return reqCtx },
		ConfigFn:         func() *raiden.Config { return cfg },
		SendJsonFn: func(data any) error {
			response = data.(controllers.AclUserPolicyResponse)
			return nil
		},
		SendErrorFn:         func(message string) error { return nil },
		SendErrorWithCodeFn: func(statusCode int, err error) error { return err },
		Data:                make(map[string]any),
	}

	controller := controllers.AclUserPolicyController{Payload: &controllers.AclUserPolicyRequest{}}

	err := controller.Get(mockCtx)
	require.NoError(t, err)
	require.Len(t, response, 2)
	require.Contains(t, response, "orders.select.orders_read")
	require.Contains(t, response, "orders.update.orders_write")
}
