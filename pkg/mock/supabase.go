package mock

import (
	"encoding/json"
	"fmt"

	"github.com/jarcoal/httpmock"
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/supabase"
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

	return registerMock(m.Cfg, "findProject", method, url, httpCode, projects)
}

func (m *MockSupabase) MockGetTablesWithExpectedResponse(httpCode int, tables []objects.Table) error {
	actionType, method, url := getMethodAndUrl(m.Cfg, "getTables")

	return registerMock(m.Cfg, actionType, method, url, httpCode, tables)
}

func (m *MockSupabase) MockGetTableByNameWithExpectedResponse(httpCode int, table objects.Table) error {
	actionType, method, url := getMethodAndUrl(m.Cfg, "getTables")

	return registerMock(m.Cfg, actionType, method, url, httpCode, []objects.Table{table})
}

func (m *MockSupabase) MockCreateTableWithExpectedResponse(httpCode int, table objects.Table) error {
	actionType, method, url := getMethodAndUrl(m.Cfg, "common")

	return registerMock(m.Cfg, actionType, method, url, httpCode, []objects.Table{table})
}

func (m *MockSupabase) MockUpdateTableWithExpectedResponse(httpCode int) error {
	actionType, method, url := getMethodAndUrl(m.Cfg, "common")

	return registerMock(m.Cfg, actionType, method, url, httpCode, objects.Table{})
}

func (m *MockSupabase) MockDeleteTableWithExpectedResponse(httpCode int) error {
	actionType, method, url := getMethodAndUrl(m.Cfg, "common")

	return registerMock(m.Cfg, actionType, method, url, httpCode, objects.Table{})
}

func (m *MockSupabase) MockGetRolesWithExpectedResponse(httpCode int, roles []objects.Role) error {
	actionType, method, url := getMethodAndUrl(m.Cfg, "getRoles")

	return registerMock(m.Cfg, actionType, method, url, httpCode, roles)
}

func (m *MockSupabase) MockGetRoleByNameWithExpectedResponse(httpCode int, role objects.Role) error {
	actionType, method, url := getMethodAndUrl(m.Cfg, "common")

	return registerMock(m.Cfg, actionType, method, url, httpCode, []objects.Role{role})
}

func (m *MockSupabase) MockCreateRoleWithExpectedResponse(httpCode int, role objects.Role) error {
	actionType, method, url := getMethodAndUrl(m.Cfg, "common")

	return registerMock(m.Cfg, actionType, method, url, httpCode, []objects.Role{role})
}

func (m *MockSupabase) MockUpdateRoleWithExpectedResponse(httpCode int) error {
	actionType, method, url := getMethodAndUrl(m.Cfg, "common")

	return registerMock(m.Cfg, actionType, method, url, httpCode, objects.Role{})
}

func (m *MockSupabase) MockDeleteRoleWithExpectedResponse(httpCode int) error {
	actionType, method, url := getMethodAndUrl(m.Cfg, "common")

	return registerMock(m.Cfg, actionType, method, url, httpCode, objects.Role{})
}

func (m *MockSupabase) MockGetPoliciesWithExpectedResponse(httpCode int, policies []objects.Policy) error {
	actionType, method, url := getMethodAndUrl(m.Cfg, "getPolicies")

	return registerMock(m.Cfg, actionType, method, url, httpCode, policies)
}

func (m *MockSupabase) MockGetPolicyByNameWithExpectedResponse(httpCode int, policy objects.Policy) error {
	actionType, method, url := getMethodAndUrl(m.Cfg, "common")

	return registerMock(m.Cfg, actionType, method, url, httpCode, []objects.Policy{policy})
}

func (m *MockSupabase) MockCreatePolicyWithExpectedResponse(httpCode int, policy objects.Policy) error {
	actionType, method, url := getMethodAndUrl(m.Cfg, "common")

	return registerMock(m.Cfg, actionType, method, url, httpCode, []objects.Policy{policy})
}

func (m *MockSupabase) MockUpdatePolicyWithExpectedResponse(httpCode int) error {
	actionType, method, url := getMethodAndUrl(m.Cfg, "common")

	return registerMock(m.Cfg, actionType, method, url, httpCode, objects.Policy{})
}

func (m *MockSupabase) MockDeletePolicyWithExpectedResponse(httpCode int) error {
	actionType, method, url := getMethodAndUrl(m.Cfg, "common")

	return registerMock(m.Cfg, actionType, method, url, httpCode, objects.Policy{})
}

func (m *MockSupabase) MockGetFunctionsWithExpectedResponse(httpCode int, functions []objects.Function) error {
	actionType, method, url := getMethodAndUrl(m.Cfg, "getFunctions")

	return registerMock(m.Cfg, actionType, method, url, httpCode, functions)
}

func (m *MockSupabase) MockGetFunctionByNameWithExpectedResponse(httpCode int, function objects.Function) error {
	actionType, method, url := getMethodAndUrl(m.Cfg, "postQuery")

	return registerMock(m.Cfg, actionType, method, url, httpCode, []objects.Function{function})
}

func (m *MockSupabase) MockCreateFunctionWithExpectedResponse(httpCode int, function objects.Function) error {
	actionType, method, url := getMethodAndUrl(m.Cfg, "postQuery")

	return registerMock(m.Cfg, actionType, method, url, httpCode, []objects.Function{function})
}

func (m *MockSupabase) MockUpdateFunctionWithExpectedResponse(httpCode int) error {
	actionType, method, url := getMethodAndUrl(m.Cfg, "postQuery")

	return registerMock(m.Cfg, actionType, method, url, httpCode, objects.Function{})
}

func (m *MockSupabase) MockDeleteFunctionWithExpectedResponse(httpCode int) error {
	actionType, method, url := getMethodAndUrl(m.Cfg, "postQuery")

	return registerMock(m.Cfg, actionType, method, url, httpCode, objects.Function{})
}

func (m *MockSupabase) MockAdminUpdateUserDataWithExpectedResponse(httpCode int, user objects.User) error {
	actionType, method, url := getMethodAndUrl(m.Cfg, "adminUpdateUserData")

	return registerMock(m.Cfg, actionType, method, url, httpCode, user)
}

func (m *MockSupabase) MockGetBucketsWithExpectedResponse(httpCode int, data interface{}) error {
	actionType, method, url := getMethodAndUrl(m.Cfg, "getBuckets")

	return registerMock(m.Cfg, actionType, method, url, httpCode, data)
}

func (m *MockSupabase) MockGetBucketByNameWithExpectedResponse(httpCode int, bucket objects.Bucket) error {
	actionType, method, url := getMethodAndUrl(m.Cfg, "getBuckets")

	return registerMock(m.Cfg, actionType, method, url+"/"+bucket.Name, httpCode, bucket)
}

func (m *MockSupabase) MockCreateBucketsWithExpectedResponse(httpCode int, data interface{}) error {
	actionType, method, url := getMethodAndUrl(m.Cfg, "createBucket")

	return registerMock(m.Cfg, actionType, method, url, httpCode, data)
}

func (m *MockSupabase) MockDeleteBucketsWithExpectedResponse(httpCode int) error {
	actionType, method, url := getMethodAndUrl(m.Cfg, "deleteBucket")

	return registerMock(m.Cfg, actionType, method, url, httpCode, supabase.DefaultBucketSuccessResponse{})
}

func registerMock(cfg *raiden.Config, actionType string, method string, url string, httpCode int, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	httpmock.RegisterResponder(method, url, httpmock.NewStringResponder(httpCode, string(jsonData)))

	if cfg.DeploymentTarget == raiden.DeploymentTargetCloud {
		switch actionType {
		case "getTables":
			httpmock.RegisterMatcherResponder(method, url,
				httpmock.BodyContainsString("join pg_namespace nsa on csa.relnamespace = nsa.oid"),
				httpmock.NewStringResponder(httpCode, string(jsonData)))
		case "getRoles":
			httpmock.RegisterMatcherResponder(method, url,
				httpmock.BodyContainsString("rolname AS name"),
				httpmock.NewStringResponder(httpCode, string(jsonData)))
		case "getPolicies":
			httpmock.RegisterMatcherResponder(method, url,
				httpmock.BodyContainsString("pol.polname AS name"),
				httpmock.NewStringResponder(httpCode, string(jsonData)))
		case "getFunctions":
			httpmock.RegisterMatcherResponder(method, url,
				httpmock.BodyContainsString("pg_namespace n on f.pronamespace = n.oid"),
				httpmock.NewStringResponder(httpCode, string(jsonData)))
		}
	}

	return nil
}

func getMethodAndUrl(cfg *raiden.Config, actionType string) (string, string, string) {
	var method string
	var url string

	action := actionType

	if cfg.DeploymentTarget == raiden.DeploymentTargetCloud {
		switch actionType {
		case "adminUpdateUserData":
			method = "PUT"
			url = fmt.Sprintf("%s/auth/v1/admin/users/user-id", cfg.SupabaseApiUrl)
		case "getBuckets":
			method = "GET"
			url = fmt.Sprintf("%s/storage/v1/bucket", cfg.SupabaseApiUrl)
		case "createBucket":
			method = "POST"
			url = fmt.Sprintf("%s/storage/v1/bucket", cfg.SupabaseApiUrl)
		case "deleteBucket":
			method = "DELETE"
			url = fmt.Sprintf("%s/storage/v1/bucket/", cfg.SupabaseApiUrl)
		default:
			method = "POST"
			url = fmt.Sprintf("%s/v1/projects/%s/database/query", cfg.SupabaseApiUrl, cfg.ProjectId)
		}
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
		case "getFunctions":
			method = "GET"
			url = fmt.Sprintf("%s%s/functions", cfg.SupabaseApiUrl, cfg.SupabaseApiBasePath)
		case "postQuery":
			method = "POST"
			url = fmt.Sprintf("%s/query", cfg.SupabaseApiUrl)
		case "getBuckets":
			method = "GET"
			url = fmt.Sprintf("%s/bucket", cfg.SupabaseApiUrl)
		case "createBucket":
			method = "POST"
			url = fmt.Sprintf("%s/bucket", cfg.SupabaseApiUrl)
		case "deleteBucket":
			method = "DELETE"
			url = fmt.Sprintf("%s/bucket", cfg.SupabaseApiUrl)
		default:
			method = "POST"
			url = fmt.Sprintf("%s%s/query", cfg.SupabaseApiUrl, cfg.SupabaseApiBasePath)
		}
	}

	return action, method, url
}
