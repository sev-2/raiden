package cloud

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/supabase/client"
	"github.com/sev-2/raiden/pkg/supabase/drivers/cloud/query"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/valyala/fasthttp"
)

func FindProject(cfg *raiden.Config) (objects.Project, error) {
	url := fmt.Sprintf("%s/v1/projects", cfg.SupabaseApiUrl)
	reqInterceptor := func(req *fasthttp.Request) error {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.AccessToken))
		return nil
	}
	projects, err := client.Get[[]objects.Project](url, client.DefaultTimeout, reqInterceptor, nil)
	if err != nil {
		return objects.Project{}, err
	}

	for i := range projects {
		p := projects[i]
		if p.Name == cfg.ProjectName {
			return p, nil
		}
	}

	return objects.Project{}, nil
}

func GetRoles(cfg *raiden.Config) ([]objects.Role, error) {
	reqInterceptor := func(req *fasthttp.Request) error {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.AccessToken))
		return nil
	}

	// response decorator
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

	resInterceptor := func(res *fasthttp.Response) error {
		var arrResponse []any
		if err := json.Unmarshal(res.Body(), &arrResponse); err != nil {
			return err
		}

		decoratedRes := resultDecoratorFn(arrResponse)
		byteData, err := json.Marshal(decoratedRes)
		if err != nil {
			return err
		}

		res.SetBodyRaw(byteData)
		return nil
	}

	rs, err := ExecuteQuery[[]objects.Role](cfg.SupabaseApiUrl, cfg.ProjectId, query.GetRolesQuery, reqInterceptor, resInterceptor)
	if err != nil {
		err = fmt.Errorf("get roles error : %s", err)
	}

	return rs, err
}

func GetPolicies(cfg *raiden.Config) ([]objects.Policy, error) {
	reqInterceptor := func(req *fasthttp.Request) error {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.AccessToken))
		return nil
	}

	rs, err := ExecuteQuery[[]objects.Policy](cfg.SupabaseApiUrl, cfg.ProjectId, query.GetPoliciesQuery, reqInterceptor, nil)
	if err != nil {
		err = fmt.Errorf("get policies error : %s", err)
	}

	return rs, err
}

func GetFunctions(cfg *raiden.Config) ([]objects.Function, error) {
	reqInterceptor := func(req *fasthttp.Request) error {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.AccessToken))
		return nil
	}

	rs, err := ExecuteQuery[[]objects.Function](cfg.SupabaseApiUrl, cfg.ProjectId, query.GenerateFunctionsQuery([]string{"public"}), reqInterceptor, nil)
	if err != nil {
		err = fmt.Errorf("get functions error : %s", err)
	}

	return rs, err
}

// ----- Execute Query -----
type ExecuteQueryParam struct {
	Query string `json:"query"`
}

func ExecuteQuery[T any](baseUrl, projectId, query string, reqInterceptor client.RequestInterceptor, resInterceptor client.ResponseInterceptor) (result T, err error) {
	url := fmt.Sprintf("%s/v1/projects/%s/database/query", baseUrl, projectId)
	p := ExecuteQueryParam{Query: query}
	pByte, err := json.Marshal(p)
	if err != nil {
		return result, err
	}

	return client.Post[T](url, pByte, client.DefaultTimeout, reqInterceptor, resInterceptor)
}
