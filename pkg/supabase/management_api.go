package supabase

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/sev-2/raiden/pkg/logger"
	management_api "github.com/sev-2/raiden/pkg/supabase/management-api"
	meta_sql "github.com/sev-2/raiden/pkg/supabase/meta-api/sql"
)

type (
	Region        string
	Project       management_api.ProjectResponse
	Organization  management_api.OrganizationResponseV1
	Organizations []management_api.OrganizationResponseV1
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
	managementApiConfiguration.Scheme = urlParsed.Scheme
	managementApiConfiguration.BasePath = ""
	// managementApiConfiguration.HTTPClient = &http.Client{
	// 	Transport: &HttpTransport{
	// 		Transport: http.DefaultTransport,
	// 	},
	// }
	managementApiConfiguration.AddDefaultHeader("Authorization", "Bearer "+accessToken)

	return management_api.NewAPIClient(managementApiConfiguration)
}

func getManagementApi() *management_api.APIClient {
	if managementApi == nil {
		managementApi = NewManagementApi()
	}
	return managementApi
}

// CustomTransport is a custom http.RoundTripper that prints request details
type HttpTransport struct {
	Transport http.RoundTripper
}

func (c *HttpTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Print the request details
	dump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		return nil, err
	}
	fmt.Println("Request:")
	fmt.Println(string(dump))

	// Use the original transport to perform the actual HTTP round trip
	resp, err := c.Transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	// Print the response details
	dump, err = httputil.DumpResponse(resp, true)
	if err != nil {
		return nil, err
	}
	fmt.Println("Response:")
	fmt.Println(string(dump))

	return resp, nil
}

func ConfigureManagementApi(url, token string) {
	apiUrl = url
	accessToken = token
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

func cloudGetTables(projectID string, includeColumn bool) (tables []Table, err error) {
	query, err := meta_sql.GenerateTablesQuery([]string{"public", "auth"}, includeColumn)
	if err != nil {
		err = fmt.Errorf("failed generate query get table for project id %s : %v", projectID, err)
		return
	}

	maybeTables, response, err := getManagementApi().ProjectsBetaApi.V1RunQuery(context.Background(), management_api.RunQueryBody{
		Query: query,
	}, projectID)

	if err != nil {
		err = fmt.Errorf("failed get table for project id %s : %v", projectID, err)
		return
	}

	if response.StatusCode != http.StatusCreated {
		err = fmt.Errorf("get all tables for project id %s got status code %v", projectID, response.StatusCode)
		return
	}

	jsonStr, err := json.Marshal(maybeTables)
	if err != nil {
		err = fmt.Errorf("invalid marshall data for get tables with project id %s : %v", projectID, err)
		return
	}

	if err = json.Unmarshal(jsonStr, &tables); err != nil {
		err = fmt.Errorf("invalid response data for get tables with project id %s : %v", projectID, err)
		return
	}

	return
}
