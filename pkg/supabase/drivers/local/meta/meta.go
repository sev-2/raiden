package meta

import (
	"encoding/json"
	"fmt"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/supabase/client"
)

// ----- Execute Query -----
type ExecuteQueryParam struct {
	Query     string `json:"query"`
	Variables any    `json:"variables"`
}

func ExecuteQuery[T any](baseUrl, query string, variables any, reqInterceptor client.RequestInterceptor, resInterceptor client.ResponseInterceptor) (result T, err error) {
	url := fmt.Sprintf("%s/query", baseUrl)
	p := ExecuteQueryParam{Query: query, Variables: variables}
	pByte, err := json.Marshal(p)
	if err != nil {
		logger.Errorf("error execute query : %s", query)
		return result, err
	}
	return client.Post[T](url, pByte, client.DefaultTimeout, reqInterceptor, resInterceptor)
}

func getBaseUrl(cfg *raiden.Config) string {
	return fmt.Sprintf("%s%s", cfg.SupabaseApiUrl, cfg.SupabaseApiBasePath)
}
