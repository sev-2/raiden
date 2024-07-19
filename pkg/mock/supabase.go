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
	data, err := json.Marshal(projects)
	if err != nil {
		return err
	}

	httpmock.RegisterResponder(method, url, httpmock.NewStringResponder(httpCode, string(data)))
	return nil
}

func (m *MockSupabase) MockGetTablesWithExpectedResponse(httpCode int, tables []objects.Table) error {
	var method = "GET"
	var url = fmt.Sprintf("%s%s/tables", m.Cfg.SupabaseApiUrl, m.Cfg.SupabaseApiBasePath)
	if m.Cfg.DeploymentTarget == raiden.DeploymentTargetCloud {
		method = "POST"
		url = fmt.Sprintf("%s/v1/projects/%s/database/query", m.Cfg.SupabaseApiUrl, m.Cfg.ProjectId)
	}

	data, err := json.Marshal(tables)
	if err != nil {
		return err
	}

	httpmock.RegisterResponder(method, url, httpmock.NewStringResponder(httpCode, string(data)))
	return nil
}

func (m *MockSupabase) MockGetTableByNameWithExpectedResponse(httpCode int, table objects.Table) error {
	var method = "GET"
	var url = fmt.Sprintf("%s%s/tables", m.Cfg.SupabaseApiUrl, m.Cfg.SupabaseApiBasePath)
	if m.Cfg.DeploymentTarget == raiden.DeploymentTargetCloud {
		method = "POST"
		url = fmt.Sprintf("%s/v1/projects/%s/database/query", m.Cfg.SupabaseApiUrl, m.Cfg.ProjectId)
	}

	data, err := json.Marshal([]objects.Table{table})
	if err != nil {
		return err
	}

	httpmock.RegisterResponder(method, url, httpmock.NewStringResponder(httpCode, string(data)))
	return nil
}

func (m *MockSupabase) MockCreateTableWithExpectedResponse(httpCode int, table objects.Table) error {
	var method = "POST"
	var url = fmt.Sprintf("%s%s/query", m.Cfg.SupabaseApiUrl, m.Cfg.SupabaseApiBasePath)
	if m.Cfg.DeploymentTarget == raiden.DeploymentTargetCloud {
		method = "POST"
		url = fmt.Sprintf("%s/v1/projects/%s/database/query", m.Cfg.SupabaseApiUrl, m.Cfg.ProjectId)
	}

	data, err := json.Marshal([]objects.Table{table})
	if err != nil {
		return err
	}

	httpmock.RegisterResponder(method, url, httpmock.NewStringResponder(httpCode, string(data)))
	return nil
}
