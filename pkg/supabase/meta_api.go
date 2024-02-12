package supabase

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/antihax/optional"
	"github.com/sev-2/raiden/pkg/logger"
	meta_api "github.com/sev-2/raiden/pkg/supabase/meta-api"
)

var (
	apiBasePath string
	metaApi     *meta_api.APIClient
)

func ConfigurationMetaApi(metaApiUrl string, metaApiBasePath string) {
	apiUrl = metaApiUrl
	apiBasePath = metaApiBasePath
}

func NewMetaApi() *meta_api.APIClient {
	urlParsed, err := url.Parse(apiUrl)
	if err != nil {
		logger.Panicf("failed parse supabase api url : %v", err)
	}

	metaApiConfiguration := meta_api.NewConfiguration()
	metaApiConfiguration.Host = urlParsed.Host
	logger.Debug("Pg-Meta Client - set host : ", metaApiConfiguration.Host)
	metaApiConfiguration.Scheme = urlParsed.Scheme
	logger.Debug("Pg-Meta Client - set schema : ", metaApiConfiguration.Scheme)
	metaApiConfiguration.BasePath = apiBasePath
	logger.Debug("Pg-Meta Client - base path  : ", metaApiConfiguration.BasePath)
	// metaApiConfiguration.HTTPClient = &LoggerHttpClient

	return meta_api.NewAPIClient(metaApiConfiguration)
}

func getMetaApi() *meta_api.APIClient {
	if metaApi == nil {
		metaApi = NewMetaApi()
	}
	return metaApi
}

func metaGetTables(ctx context.Context, includedSchema []string, includeColumns bool) (tables []Table, err error) {
	var includedSchemaParam string
	if len(includedSchema) > 0 {
		includedSchemaParam = strings.Join(includedSchema, ",")
	}

	data, response, err := getMetaApi().DefaultApi.TablesGet(ctx, "", &meta_api.DefaultApiTablesGetOpts{
		IncludedSchemas: optional.NewString(includedSchemaParam),
		IncludeColumns:  optional.NewBool(includeColumns),
	})
	if err != nil {
		return tables, fmt.Errorf("failed get all table : %s", err)
	}
	return marshallResponse[[]Table]("get all table", data, response)
}

func metaGetRoles(ctx context.Context) (roles []Role, err error) {
	data, response, err := getMetaApi().DefaultApi.RolesGet(ctx, "", &meta_api.DefaultApiRolesGetOpts{})
	if err != nil {
		return roles, fmt.Errorf("failed get all role : %s", err)
	}
	return marshallResponse[[]Role]("get all role", data, response)
}

func metaGetPolicies(ctx context.Context) (policies []Policy, err error) {
	data, response, err := getMetaApi().DefaultApi.PoliciesGet(ctx)
	if err != nil {
		return policies, fmt.Errorf("failed get all policy : %s", err)
	}
	return marshallResponse[[]Policy]("get all policy", data, response)
}

func metaGetFunctions(ctx context.Context) (functions []Function, err error) {
	data, response, err := getMetaApi().DefaultApi.FunctionsGet(ctx)
	if err != nil {
		return functions, fmt.Errorf("failed get all functions : %s", err)
	}
	return marshallResponse[[]Function]("get all functions", data, response)
}

func metaExecQuery(ctx context.Context, query string) (result any, err error) {
	data, response, err := getMetaApi().DefaultApi.QueryPost(ctx, meta_api.QueryBody{
		Query: query,
	})

	if err != nil {
		err = fmt.Errorf("failed execute query : %v", err)
		return
	}

	if response.StatusCode >= 400 {
		err = fmt.Errorf("execute query got status code %v", response.StatusCode)
		return
	}

	return data, nil
}
