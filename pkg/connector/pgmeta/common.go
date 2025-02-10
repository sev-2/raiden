package pgmeta

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sev-2/raiden/pkg/client/net"
	"github.com/sev-2/raiden/pkg/logger"
)

var MetaLogger = logger.HcLog().Named("connector.meta")

func DefaultAuthInterceptor(accessToken string) func(req *http.Request) error {
	return func(req *http.Request) error {
		if accessToken != "" {
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
		}
		return nil
	}
}

// ----- Execute Query -----
type ExecuteQueryParam struct {
	Query     string `json:"query"`
	Variables any    `json:"variables"`
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
