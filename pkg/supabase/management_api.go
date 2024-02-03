package supabase

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/sev-2/raiden/pkg/logger"
	management_api "github.com/sev-2/raiden/pkg/supabase/management-api"
	meta_sql "github.com/sev-2/raiden/pkg/supabase/meta-api/sql"
)

type (
	Region        string
	Project       management_api.ProjectResponse
	Organization  management_api.OrganizationResponseV1
	Organizations []management_api.OrganizationResponseV1
	ApiKeys       []management_api.ApiKeyResponse
)

var (
	accessToken   string
	managementApi *management_api.APIClient
	AllRegion     []Region = []Region{
		RegionSoutheastAsia,
		RegionNortheastAsia,
		RegionNortheastAsia2,
		RegionCentralCanada,
		RegionWestUS,
		RegionEastUS,
		RegionWestEU,
		RegionWestEU2,
		RegionCentralEU,
		RegionSouthAsia,
		RegionOceania,
		RegionSouthAmerica,
	}
)

const (
	RegionSoutheastAsia  Region = "ap-southeast-1"
	RegionNortheastAsia  Region = "ap-northeast-1"
	RegionNortheastAsia2 Region = "ap-northeast-2"
	RegionCentralCanada  Region = "ca-central-1"
	RegionWestUS         Region = "us-west-1"
	RegionEastUS         Region = "es-east-1"
	RegionWestEU         Region = "eu-west-1"
	RegionWestEU2        Region = "eu-west-2"
	RegionCentralEU      Region = "eu-central-1"
	RegionSouthAsia      Region = "ap-south-1"
	RegionOceania        Region = "ap-southeast-2"
	RegionSouthAmerica   Region = "sa-east-1"
)

func NewManagementApi() *management_api.APIClient {
	urlParsed, err := url.Parse(apiUrl)
	if err != nil {
		logger.Panicf("failed parse supabase api url : %v", err)
	}

	managementApiConfiguration := management_api.NewConfiguration()
	managementApiConfiguration.Host = urlParsed.Host
	logger.Debug("Supabase Client - set host : ", managementApiConfiguration.Host)
	managementApiConfiguration.Scheme = urlParsed.Scheme
	logger.Debug("Supabase Client - set schema : ", managementApiConfiguration.Scheme)
	managementApiConfiguration.BasePath = ""
	logger.Debug("Supabase Client - set base path : ", managementApiConfiguration.BasePath)
	// managementApiConfiguration.HTTPClient = &LoggerHttpClient

	managementApiConfiguration.AddDefaultHeader("Authorization", "Bearer "+accessToken)
	logger.Debug("Supabase Client - set authorization token : ", accessToken)

	return management_api.NewAPIClient(managementApiConfiguration)
}

func getManagementApi() *management_api.APIClient {
	if managementApi == nil {
		managementApi = NewManagementApi()
	}
	return managementApi
}

func ConfigureManagementApi(url, token string) {
	apiUrl = url
	accessToken = token
}

func GetProjectKeys(projectId string) (keys ApiKeys, err error) {
	keys, response, err := getManagementApi().ProjectsApi.GetProjectApiKeys(context.Background(), projectId)
	if err != nil {
		err = fmt.Errorf("failed get project keys : %v", err)
		return
	}

	if response.StatusCode != http.StatusOK {
		err = fmt.Errorf("get project keys got status code %v", response.StatusCode)
		return
	}

	return keys, nil

}

func FindProject(projectName string) (project Project, err error) {
	projects, response, err := getManagementApi().ProjectsApi.GetProjects(context.Background())
	if err != nil {
		err = fmt.Errorf("failed check project %s is exist : %v", projectName, err)
		return
	}

	if response.StatusCode != http.StatusOK {
		err = fmt.Errorf("check project %s got status code %v", projectName, response.StatusCode)
		return
	}

	for _, v := range projects {
		if v.Name == projectName {
			project = Project(v)
			return
		}
	}

	return project, nil
}

func CreateNewProject(data management_api.CreateProjectBody) (Project, error) {
	dataByte, err := json.Marshal(data)
	if err != nil {
		return Project{}, err
	}
	logger.Debugf("CreateNewProject - with payload : %s", string(dataByte))

	project, response, err := getManagementApi().ProjectsApi.CreateProject(context.Background(), data)
	if err != nil {
		err := fmt.Errorf("failed create new project %s : %v", data.Name, err)
		return Project(project), err
	}

	if response.StatusCode != http.StatusCreated {
		err := fmt.Errorf("create new project %s got status code %v", data.Name, response.StatusCode)
		return Project(project), err
	}

	return Project(project), nil
}

func GetOrganizations() (Organizations, error) {
	organizations, response, err := getManagementApi().OrganizationsApi.GetOrganizations(context.Background())
	if err != nil {
		err := fmt.Errorf("failed get organizations : %v", err)
		return organizations, err
	}

	if response.StatusCode != http.StatusOK {
		err := fmt.Errorf("get organizations got status code %v", response.StatusCode)
		return organizations, err
	}

	return organizations, nil
}

func getTables(ctx context.Context, projectId string, includeColumn bool) (tables []Table, err error) {
	query, err := meta_sql.GenerateTablesQuery(DefaultIncludedSchema, includeColumn)
	if err != nil {
		err = fmt.Errorf("failed generate query get table for project id %s : %v", projectId, err)
		return
	}

	return executeQuery[[]Table](ctx, projectId, "get tables", query, nil)
}

func getRoles(ctx context.Context, projectId string) (roles []Role, err error) {
	findConfigFn := func(role any) []any {
		if roleMap, isMapAny := role.(map[string]any); isMapAny {
			if configValue, exist := roleMap["config"]; exist {
				if configArr, isArrayAny := configValue.([]any); isArrayAny {
					return configArr
				}
			}
		}

		return nil
	}

	configsToMapFn := func(configs []any) map[string]any {
		mapConfig := make(map[string]any)
		for _, configItem := range configs {
			if configItemStr, isString := configItem.(string); isString {
				configItemSplitted := strings.Split(configItemStr, "=")
				if len(configItemSplitted) == 2 {
					mapConfig[configItemSplitted[0]] = configItemSplitted[1]
				}
			}
		}
		return mapConfig
	}

	resultDecoratorFn := func(result any) any {
		if roles, isRolesArr := result.([]any); isRolesArr {
			for roleIndex := range roles {
				roleItem := roles[roleIndex]
				if foundConfig := findConfigFn(roleItem); foundConfig != nil {
					config := configsToMapFn(foundConfig)
					if config != nil {
						roleItem.(map[string]any)["config"] = config
					}
				}
			}
		}
		return result
	}
	return executeQuery[[]Role](ctx, projectId, "get roles", meta_sql.GetRolesQuery, resultDecoratorFn)
}

func getPolicies(ctx context.Context, projectId string) (policies []Policy, err error) {
	return executeQuery[[]Policy](ctx, projectId, "get policies", meta_sql.GetPoliciesQuery, nil)
}

func getFunctions(ctx context.Context, projectId string) (functions []Function, err error) {
	return executeQuery[[]Function](ctx, projectId, "get functions", meta_sql.GetFunctionsQuery, nil)
}

func executeQuery[T any](ctx context.Context, projectId, action, query string, resultDecorator func(response any) any) (result T, err error) {
	anyResult, response, err := getManagementApi().ProjectsBetaApi.V1RunQuery(ctx, management_api.RunQueryBody{
		Query: query,
	}, projectId)

	if err != nil {
		err = fmt.Errorf("failed %s for project id %s : %v", action, projectId, err)
		return
	}

	if resultDecorator != nil {
		anyResult = resultDecorator(anyResult)
	}

	if response.StatusCode != http.StatusCreated {
		err = fmt.Errorf("%s for project id %s got status code %v", action, projectId, response.StatusCode)
		return
	}

	jsonStr, err := json.Marshal(anyResult)
	if err != nil {
		err = fmt.Errorf("invalid marshall data for %s with project id %s : %v", action, projectId, err)
		return
	}

	if err = json.Unmarshal(jsonStr, &result); err != nil {
		err = fmt.Errorf("invalid response data for %s with project id %s : %v", action, projectId, err)
		return
	}
	return
}
