package cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/supabase/client/net"
	"github.com/sev-2/raiden/pkg/supabase/objects"
)

var CloudLogger = logger.HcLog().Named("supabase.cloud")

func DefaultAuthInterceptor(accessToken string) func(req *http.Request) error {
	return func(req *http.Request) error {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
		return nil
	}
}

func FindProject(cfg *raiden.Config) (objects.Project, error) {
	CloudLogger.Trace("start find project from supabase")
	url := fmt.Sprintf("%s/v1/projects", cfg.SupabaseApiUrl)
	projects, err := net.Get[[]objects.Project](url, net.DefaultTimeout, DefaultAuthInterceptor(cfg.AccessToken), nil)
	if err != nil {
		return objects.Project{}, err
	}

	for i := range projects {
		p := projects[i]
		if p.Name == cfg.ProjectName {
			return p, nil
		}
	}
	CloudLogger.Trace("finish find project from supabase")
	return objects.Project{}, nil
}

// ----- Execute Query -----
type ExecuteQueryParam struct {
	Query string `json:"query"`
}

func ExecuteQuery[T any](baseUrl, projectId, query string, reqInterceptor net.RequestInterceptor, resInterceptor net.ResponseInterceptor) (result T, err error) {
	url := fmt.Sprintf("%s/v1/projects/%s/database/query", baseUrl, projectId)
	p := ExecuteQueryParam{Query: query}
	pByte, err := json.Marshal(p)
	if err != nil {
		CloudLogger.Error("error execute query", "query", query)
		return result, err
	}

	return net.Post[T](url, pByte, net.DefaultTimeout, reqInterceptor, resInterceptor)
}

func cleanupQueryParam(q string) string {
	cleanQuery := strings.ReplaceAll(q, "\n", " ")
	cleanQuery = strings.ReplaceAll(cleanQuery, "\t", " ")
	cleanQuery = strings.ReplaceAll(cleanQuery, "  ", "")
	return cleanQuery
}
