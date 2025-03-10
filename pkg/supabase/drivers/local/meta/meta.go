package meta

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/client/net"
	"github.com/sev-2/raiden/pkg/logger"
)

var MetaLogger = logger.HcLog().Named("supabase.meta")

// ----- Execute Query -----
type ExecuteQueryParam struct {
	Query     string `json:"query"`
	Variables any    `json:"variables"`
}

func DefaultInterceptor(cfg *raiden.Config) net.RequestInterceptor {
	return func(req *http.Request) error {
		if cfg.SupabaseApiToken != "" {
			token := fmt.Sprintf("%s %s", cfg.SupabaseApiTokenType, cfg.SupabaseApiToken)
			req.Header.Set("Authorization", token)
		}

		return nil
	}
}

func ExecuteQuery[T any](baseUrl, query string, variables any, reqInterceptor net.RequestInterceptor, resInterceptor net.ResponseInterceptor) (result T, err error) {
	url := fmt.Sprintf("%s/query", baseUrl)
	p := ExecuteQueryParam{Query: query, Variables: variables}
	pByte, err := json.Marshal(p)
	if err != nil {
		MetaLogger.Error("error execute query", "query", query)
		return result, err
	}
	return net.Post[T](url, pByte, net.DefaultTimeout, reqInterceptor, resInterceptor)
}

func getBaseUrl(cfg *raiden.Config) string {
	return fmt.Sprintf("%s%s", cfg.SupabaseApiUrl, cfg.SupabaseApiBasePath)
}
