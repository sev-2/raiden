package cloud

import (
	"encoding/json"
	"fmt"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/supabase/client"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/valyala/fasthttp"
)

func DefaultAuthInterceptor(accessToken string) func(req *fasthttp.Request) error {
	return func(req *fasthttp.Request) error {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
		return nil
	}
}

func FindProject(cfg *raiden.Config) (objects.Project, error) {
	url := fmt.Sprintf("%s/v1/projects", cfg.SupabaseApiUrl)
	projects, err := client.Get[[]objects.Project](url, client.DefaultTimeout, DefaultAuthInterceptor(cfg.AccessToken), nil)
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

// ----- Execute Query -----
type ExecuteQueryParam struct {
	Query string `json:"query"`
}

func ExecuteQuery[T any](baseUrl, projectId, query string, reqInterceptor client.RequestInterceptor, resInterceptor client.ResponseInterceptor) (result T, err error) {
	url := fmt.Sprintf("%s/v1/projects/%s/database/query", baseUrl, projectId)
	p := ExecuteQueryParam{Query: query}
	pByte, err := json.Marshal(p)
	if err != nil {
		logger.Errorf("error execute query : %s", query)
		return result, err
	}

	return client.Post[T](url, pByte, client.DefaultTimeout, reqInterceptor, resInterceptor)
}