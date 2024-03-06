package meta

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/supabase/client"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/valyala/fasthttp"
)

func GetTables(cfg *raiden.Config, includedSchemas []string, includeColumns bool) ([]objects.Table, error) {
	url := fmt.Sprintf("%s%s/tables", cfg.SupabaseApiUrl, cfg.SupabaseApiBasePath)
	reqInterceptor := func(req *fasthttp.Request) error {
		if len(includedSchemas) > 0 {
			req.URI().QueryArgs().Set("included_schemas", strings.Join(includedSchemas, ","))
		}

		if includeColumns {
			req.URI().QueryArgs().Set("include_columns", strconv.FormatBool(includeColumns))
		}

		return nil
	}

	rs, err := client.Get[[]objects.Table](url, client.DefaultTimeout, reqInterceptor, nil)
	if err != nil {
		err = fmt.Errorf("get tables error : %s", err)
	}
	return rs, err
}

func GetRoles(cfg *raiden.Config) ([]objects.Role, error) {
	url := fmt.Sprintf("%s%s/roles", cfg.SupabaseApiUrl, cfg.SupabaseApiBasePath)
	rs, err := client.Get[[]objects.Role](url, client.DefaultTimeout, nil, nil)
	if err != nil {
		err = fmt.Errorf("get roles error : %s", err)
	}
	return rs, err
}

func GetPolicies(cfg *raiden.Config) ([]objects.Policy, error) {
	url := fmt.Sprintf("%s%s/policies", cfg.SupabaseApiUrl, cfg.SupabaseApiBasePath)
	rs, err := client.Get[[]objects.Policy](url, client.DefaultTimeout, nil, nil)
	if err != nil {
		err = fmt.Errorf("get roles error : %s", err)
	}
	return rs, err
}

func GetFunctions(cfg *raiden.Config) ([]objects.Function, error) {
	url := fmt.Sprintf("%s%s/functions", cfg.SupabaseApiUrl, cfg.SupabaseApiBasePath)
	rs, err := client.Get[[]objects.Function](url, client.DefaultTimeout, nil, nil)
	if err != nil {
		err = fmt.Errorf("get roles error : %s", err)
	}
	return rs, err
}
