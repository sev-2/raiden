package meta

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/client/net"
	"github.com/sev-2/raiden/pkg/supabase/objects"
)

func GetTypes(cfg *raiden.Config, includedSchemas []string) ([]objects.Type, error) {
	MetaLogger.Trace("start fetching types from meta")
	url := fmt.Sprintf("%s%s/types", cfg.SupabaseApiUrl, cfg.SupabaseApiBasePath)
	reqInterceptor := func(req *http.Request) error {
		if len(includedSchemas) > 0 {
			req.URL.Query().Set("included_schemas", strings.Join(includedSchemas, ","))
		}

		return nil
	}
	rs, err := net.Get[[]objects.Type](url, net.DefaultTimeout, reqInterceptor, nil)
	if err != nil {
		err = fmt.Errorf("get types error : %s", err)
	}
	MetaLogger.Trace("finish fetching types from meta")
	return rs, err
}
