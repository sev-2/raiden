package mock

import (
	"encoding/json"
	"fmt"

	"github.com/jarcoal/httpmock"
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/supabase/objects"
)

type MockSupabase struct {
	Cfg *raiden.Config
}

func (m *MockSupabase) Activate() {
	httpmock.Activate()
}

func (m *MockSupabase) Deactivate() {
	defer httpmock.DeactivateAndReset()
}

func (m *MockSupabase) MockFindProjectWithExpectedResponse(httpCode int, project objects.Project) error {
	if m.Cfg.DeploymentTarget == raiden.DeploymentTargetSelfHosted {
		return fmt.Errorf("FindProject not implemented for self hosted")
	}

	var method = "GET"
	var url = fmt.Sprintf("%s%s/projects", m.Cfg.SupabaseApiUrl, m.Cfg.SupabaseApiBasePath)
	projects := []objects.Project{project}

	return registerMock(method, url, httpCode, projects)
}

func (m *MockSupabase) MockGetTablesWithExpectedResponse(httpCode int, tables []objects.Table) error {
	method, url := getMethodAndUrl(m.Cfg, "getTables")

	return registerMock(method, url, httpCode, tables)
}

func (m *MockSupabase) MockGetTableByNameWithExpectedResponse(httpCode int, table objects.Table) error {
	method, url := getMethodAndUrl(m.Cfg, "getTables")

	return registerMock(method, url, httpCode, []objects.Table{table})
}

func (m *MockSupabase) MockCreateTableWithExpectedResponse(httpCode int, table objects.Table) error {
	method, url := getMethodAndUrl(m.Cfg, "common")

	return registerMock(method, url, httpCode, []objects.Table{table})
}

func (m *MockSupabase) MockUpdateTableWithExpectedResponse(httpCode int) error {
	method, url := getMethodAndUrl(m.Cfg, "common")

	return registerMock(method, url, httpCode, objects.Table{})
}

func (m *MockSupabase) MockDeleteTableWithExpectedResponse(httpCode int) error {
	method, url := getMethodAndUrl(m.Cfg, "common")

	return registerMock(method, url, httpCode, objects.Table{})
}

func (m *MockSupabase) MockGetRolesWithExpectedResponse(httpCode int, roles []objects.Role) error {
	method, url := getMethodAndUrl(m.Cfg, "getRoles")

	return registerMock(method, url, httpCode, roles)
}

func (m *MockSupabase) MockGetRoleByNameWithExpectedResponse(httpCode int, role objects.Role) error {
	method, url := getMethodAndUrl(m.Cfg, "common")

	return registerMock(method, url, httpCode, []objects.Role{role})
}

func (m *MockSupabase) MockCreateRoleWithExpectedResponse(httpCode int, role objects.Role) error {
	method, url := getMethodAndUrl(m.Cfg, "common")

	return registerMock(method, url, httpCode, []objects.Role{role})
}

func (m *MockSupabase) MockUpdateRoleWithExpectedResponse(httpCode int) error {
	method, url := getMethodAndUrl(m.Cfg, "common")

	return registerMock(method, url, httpCode, objects.Role{})
}

func (m *MockSupabase) MockDeleteRoleWithExpectedResponse(httpCode int) error {
	method, url := getMethodAndUrl(m.Cfg, "common")

	return registerMock(method, url, httpCode, objects.Role{})
}

func (m *MockSupabase) MockGetPoliciesWithExpectedResponse(httpCode int, policies []objects.Policy) error {
	method, url := getMethodAndUrl(m.Cfg, "getPolicies")

	return registerMock(method, url, httpCode, policies)
}

func (m *MockSupabase) MockCreatePolicyWithExpectedResponse(httpCode int, policy objects.Policy) error {
	method, url := getMethodAndUrl(m.Cfg, "common")

	return registerMock(method, url, httpCode, []objects.Policy{policy})
}

func (m *MockSupabase) MockUpdatePolicyWithExpectedResponse(httpCode int) error {
	method, url := getMethodAndUrl(m.Cfg, "common")

	return registerMock(method, url, httpCode, objects.Policy{})
}

func (m *MockSupabase) MockDeletePolicyWithExpectedResponse(httpCode int) error {
	method, url := getMethodAndUrl(m.Cfg, "common")

	return registerMock(method, url, httpCode, objects.Policy{})
}

func registerMock(method, url string, httpCode int, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	httpmock.RegisterResponder(method, url, httpmock.NewStringResponder(httpCode, string(jsonData)))
	return nil
}

func getMethodAndUrl(cfg *raiden.Config, actionType string) (string, string) {
	var method string
	var url string

	if cfg.DeploymentTarget == raiden.DeploymentTargetCloud {
		method = "POST"
		url = fmt.Sprintf("%s/v1/projects/%s/database/query", cfg.SupabaseApiUrl, cfg.ProjectId)
	} else {
		switch actionType {
		case "getTables":
			method = "GET"
			url = fmt.Sprintf("%s%s/tables", cfg.SupabaseApiUrl, cfg.SupabaseApiBasePath)
		case "getRoles":
			method = "GET"
			url = fmt.Sprintf("%s%s/roles", cfg.SupabaseApiUrl, cfg.SupabaseApiBasePath)
		case "getPolicies":
			method = "GET"
			url = fmt.Sprintf("%s%s/policies", cfg.SupabaseApiUrl, cfg.SupabaseApiBasePath)
		default:
			method = "POST"
			url = fmt.Sprintf("%s%s/query", cfg.SupabaseApiUrl, cfg.SupabaseApiBasePath)
		}
	}

	return method, url
}
