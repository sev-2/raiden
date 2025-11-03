package meta_test

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/sev-2/raiden"
	meta "github.com/sev-2/raiden/pkg/supabase/drivers/local/meta"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestConfig() *raiden.Config {
	return &raiden.Config{
		SupabaseApiUrl:       "http://meta.test",
		SupabaseApiBasePath:  "/pg",
		SupabaseApiTokenType: "Bearer",
		SupabaseApiToken:     "token",
	}
}

func rolesURL(cfg *raiden.Config) string {
	return fmt.Sprintf("%s%s/roles", cfg.SupabaseApiUrl, cfg.SupabaseApiBasePath)
}

func queryURL(cfg *raiden.Config) string {
	return fmt.Sprintf("%s%s/query", cfg.SupabaseApiUrl, cfg.SupabaseApiBasePath)
}

func setupHTTPMock(t *testing.T) {
	httpmock.Activate()
	t.Cleanup(httpmock.DeactivateAndReset)
}

func TestGetRolesSuccess(t *testing.T) {
	cfg := newTestConfig()
	setupHTTPMock(t)

	httpmock.RegisterResponder(http.MethodGet, rolesURL(cfg), func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, "Bearer token", req.Header.Get("Authorization"))
		return httpmock.NewStringResponse(200, `[
            {"name":"role1","config":{"search_path":"public"}},
            {"name":"role2"}
        ]`), nil
	})

	roles, err := meta.GetRoles(cfg)
	require.NoError(t, err)
	require.Len(t, roles, 2)
	assert.Equal(t, "role1", roles[0].Name)
	assert.Equal(t, "public", roles[0].Config["search_path"])
	assert.Equal(t, "role2", roles[1].Name)
}

func TestGetRolesServerError(t *testing.T) {
	cfg := newTestConfig()
	setupHTTPMock(t)

	httpmock.RegisterResponder(http.MethodGet, rolesURL(cfg), httpmock.NewStringResponder(500, `{}`))

	roles, err := meta.GetRoles(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get roles error")
	assert.Empty(t, roles)
}

func TestGetRoleMembershipsSuccess(t *testing.T) {
	cfg := newTestConfig()
	setupHTTPMock(t)

	httpmock.RegisterResponder(http.MethodPost, queryURL(cfg), func(req *http.Request) (*http.Response, error) {
		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)
		assert.Contains(t, string(body), "public")
		assert.Contains(t, string(body), "private")
		return httpmock.NewStringResponse(200, `[
            {"parent_id":1,"parent_role":"parent","inherit_id":2,"inherit_role":"child"}
        ]`), nil
	})

	memberships, err := meta.GetRoleMemberships(cfg, []string{"public", "private"})
	require.NoError(t, err)
	require.Len(t, memberships, 1)
	assert.Equal(t, "child", memberships[0].InheritRole)
}

func TestGetRoleMembershipsServerError(t *testing.T) {
	cfg := newTestConfig()
	setupHTTPMock(t)

	httpmock.RegisterResponder(http.MethodPost, queryURL(cfg), httpmock.NewStringResponder(500, `{}`))

	memberships, err := meta.GetRoleMemberships(cfg, []string{"public"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get role membership error")
	assert.Empty(t, memberships)
}

func TestGetRoleByNameSuccess(t *testing.T) {
	cfg := newTestConfig()
	setupHTTPMock(t)

	httpmock.RegisterResponder(http.MethodPost, queryURL(cfg), func(req *http.Request) (*http.Response, error) {
		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)
		assert.Contains(t, string(body), "rolname = 'role1'")
		return httpmock.NewStringResponse(200, `[
            {"name":"role1","config":{"search_path":"public"}}
        ]`), nil
	})

	role, err := meta.GetRoleByName(cfg, "role1")
	require.NoError(t, err)
	assert.Equal(t, "role1", role.Name)
	assert.Equal(t, "public", role.Config["search_path"])
}

func TestGetRoleByNameNotFound(t *testing.T) {
	cfg := newTestConfig()
	setupHTTPMock(t)

	httpmock.RegisterResponder(http.MethodPost, queryURL(cfg), httpmock.NewStringResponder(200, `[]`))

	role, err := meta.GetRoleByName(cfg, "missing")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "is not found")
	assert.Equal(t, objects.Role{}, role)
}

func TestGetRoleByNameServerError(t *testing.T) {
	cfg := newTestConfig()
	setupHTTPMock(t)

	httpmock.RegisterResponder(http.MethodPost, queryURL(cfg), httpmock.NewStringResponder(500, `{}`))

	role, err := meta.GetRoleByName(cfg, "missing")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get role error")
	assert.Equal(t, objects.Role{}, role)
}

func TestCreateRoleSuccess(t *testing.T) {
	cfg := newTestConfig()
	setupHTTPMock(t)

	var calls int32
	httpmock.RegisterResponder(http.MethodPost, queryURL(cfg), func(req *http.Request) (*http.Response, error) {
		call := atomic.AddInt32(&calls, 1)
		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)
		if call == 1 {
			assert.Contains(t, string(body), "CREATE ROLE")
			return httpmock.NewStringResponse(200, `{}`), nil
		}
		assert.Contains(t, string(body), "rolname = 'tester'")
		return httpmock.NewStringResponse(200, `[
            {"name":"tester","config":{"search_path":"admin"}}
        ]`), nil
	})

	role, err := meta.CreateRole(cfg, objects.Role{Name: "tester", Config: map[string]any{"search_path": "admin"}})
	require.NoError(t, err)
	assert.Equal(t, "tester", role.Name)
	assert.Equal(t, "admin", role.Config["search_path"])
	assert.Equal(t, int32(2), atomic.LoadInt32(&calls))
}

func TestCreateRoleServerError(t *testing.T) {
	cfg := newTestConfig()
	setupHTTPMock(t)

	httpmock.RegisterResponder(http.MethodPost, queryURL(cfg), httpmock.NewStringResponder(500, `{}`))

	role, err := meta.CreateRole(cfg, objects.Role{Name: "tester"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "create new role")
	assert.Equal(t, objects.Role{}, role)
}

func TestDeleteRoleSuccess(t *testing.T) {
	cfg := newTestConfig()
	setupHTTPMock(t)

	httpmock.RegisterResponder(http.MethodPost, queryURL(cfg), httpmock.NewStringResponder(200, `{}`))

	err := meta.DeleteRole(cfg, objects.Role{Name: "tester"})
	assert.NoError(t, err)
}

func TestDeleteRoleServerError(t *testing.T) {
	cfg := newTestConfig()
	setupHTTPMock(t)

	httpmock.RegisterResponder(http.MethodPost, queryURL(cfg), httpmock.NewStringResponder(500, `{}`))

	err := meta.DeleteRole(cfg, objects.Role{Name: "tester"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delete role")
}

func TestUpdateRoleNoChanges(t *testing.T) {
	cfg := newTestConfig()

	err := meta.UpdateRole(cfg, objects.Role{}, objects.UpdateRoleParam{OldData: objects.Role{Name: "tester"}})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "has no changes")
}

func TestUpdateRoleSuccess(t *testing.T) {
	cfg := newTestConfig()
	setupHTTPMock(t)

	var updateCalls int32
	var inheritCalls int32

	httpmock.RegisterResponder(http.MethodPost, queryURL(cfg), func(req *http.Request) (*http.Response, error) {
		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)
		if strings.Contains(string(body), "GRANT") {
			atomic.AddInt32(&inheritCalls, 1)
		} else {
			atomic.AddInt32(&updateCalls, 1)
		}
		return httpmock.NewStringResponse(200, `{}`), nil
	})

	err := meta.UpdateRole(cfg, objects.Role{Name: ""}, objects.UpdateRoleParam{
		OldData:     objects.Role{Name: "tester"},
		ChangeItems: []objects.UpdateRoleType{objects.UpdateRoleConfig},
		ChangeInheritItems: []objects.UpdateRoleInheritItem{
			{Role: objects.Role{Name: "reader"}, Type: objects.UpdateRoleInheritGrant},
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&updateCalls))
	assert.Equal(t, int32(1), atomic.LoadInt32(&inheritCalls))
}

func TestUpdateRoleInheritanceOnly(t *testing.T) {
	cfg := newTestConfig()
	setupHTTPMock(t)

	var inheritCalls int32
	httpmock.RegisterResponder(http.MethodPost, queryURL(cfg), func(req *http.Request) (*http.Response, error) {
		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)
		assert.Contains(t, string(body), "GRANT")
		atomic.AddInt32(&inheritCalls, 1)
		return httpmock.NewStringResponse(200, `{}`), nil
	})

	err := meta.UpdateRole(cfg, objects.Role{}, objects.UpdateRoleParam{
		OldData: objects.Role{Name: "tester"},
		ChangeInheritItems: []objects.UpdateRoleInheritItem{
			{Role: objects.Role{Name: "reader"}, Type: objects.UpdateRoleInheritGrant},
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&inheritCalls))
}

func TestUpdateRoleSkipsEmptyInheritance(t *testing.T) {
	cfg := newTestConfig()
	setupHTTPMock(t)

	var inheritCalls int32
	httpmock.RegisterResponder(http.MethodPost, queryURL(cfg), func(req *http.Request) (*http.Response, error) {
		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)
		if strings.Contains(string(body), "GRANT") {
			atomic.AddInt32(&inheritCalls, 1)
		}
		return httpmock.NewStringResponse(200, `{}`), nil
	})

	err := meta.UpdateRole(cfg, objects.Role{}, objects.UpdateRoleParam{
		OldData:     objects.Role{Name: "tester"},
		ChangeItems: []objects.UpdateRoleType{objects.UpdateRoleCanLogin},
		ChangeInheritItems: []objects.UpdateRoleInheritItem{
			{Role: objects.Role{Name: ""}, Type: objects.UpdateRoleInheritGrant},
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, int32(0), atomic.LoadInt32(&inheritCalls))
}

func TestUpdateRoleInheritanceInvalidType(t *testing.T) {
	cfg := newTestConfig()
	setupHTTPMock(t)

	var updateCalls int32
	httpmock.RegisterResponder(http.MethodPost, queryURL(cfg), func(req *http.Request) (*http.Response, error) {
		atomic.AddInt32(&updateCalls, 1)
		return httpmock.NewStringResponse(200, `{}`), nil
	})

	err := meta.UpdateRole(cfg, objects.Role{}, objects.UpdateRoleParam{
		OldData:     objects.Role{Name: "tester"},
		ChangeItems: []objects.UpdateRoleType{objects.UpdateRoleConfig},
		ChangeInheritItems: []objects.UpdateRoleInheritItem{
			{Role: objects.Role{Name: "reader"}, Type: objects.UpdateRoleInheritType("invalid")},
		},
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported role inherit action")
	assert.Equal(t, int32(1), atomic.LoadInt32(&updateCalls))
}

func TestUpdateRoleInheritanceExecuteError(t *testing.T) {
	cfg := newTestConfig()
	setupHTTPMock(t)

	httpmock.RegisterResponder(http.MethodPost, queryURL(cfg), func(req *http.Request) (*http.Response, error) {
		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)
		if strings.Contains(string(body), "GRANT") {
			return httpmock.NewStringResponse(500, `{}`), nil
		}
		return httpmock.NewStringResponse(200, `{}`), nil
	})

	err := meta.UpdateRole(cfg, objects.Role{}, objects.UpdateRoleParam{
		OldData:     objects.Role{Name: "tester"},
		ChangeItems: []objects.UpdateRoleType{objects.UpdateRoleConfig},
		ChangeInheritItems: []objects.UpdateRoleInheritItem{
			{Role: objects.Role{Name: "reader"}, Type: objects.UpdateRoleInheritGrant},
		},
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "grant role reader for tester error")
}

func TestUpdateRoleAggregatedErrors(t *testing.T) {
	cfg := newTestConfig()
	setupHTTPMock(t)

	httpmock.RegisterResponder(http.MethodPost, queryURL(cfg), httpmock.NewStringResponder(500, `{}`))

	err := meta.UpdateRole(cfg, objects.Role{}, objects.UpdateRoleParam{
		OldData:     objects.Role{Name: "tester"},
		ChangeItems: []objects.UpdateRoleType{objects.UpdateRoleConfig},
		ChangeInheritItems: []objects.UpdateRoleInheritItem{
			{Role: objects.Role{Name: "first"}, Type: objects.UpdateRoleInheritGrant},
			{Role: objects.Role{Name: "second"}, Type: objects.UpdateRoleInheritRevoke},
		},
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update role tester error")
	assert.Contains(t, err.Error(), "grant role first for tester error")
	assert.Contains(t, err.Error(), "revoke role second for tester error")
}
