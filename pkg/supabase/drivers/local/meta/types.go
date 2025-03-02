package meta

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/client/net"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/supabase/query"
	"github.com/sev-2/raiden/pkg/supabase/query/sql"
)

func GetTypes(cfg *raiden.Config, includedSchemas []string) ([]objects.Type, error) {
	MetaLogger.Trace("start fetching types from meta")
	url := fmt.Sprintf("%s%s/types", cfg.SupabaseApiUrl, cfg.SupabaseApiBasePath)
	reqInterceptor := func(req *http.Request) error {
		if len(includedSchemas) > 0 {
			reqQuery := req.URL.Query()
			reqQuery.Set("included_schemas", strings.Join(includedSchemas, ","))
			req.URL.RawQuery = reqQuery.Encode()
		}

		if cfg.SupabaseApiToken != "" {
			token := fmt.Sprintf("%s %s", cfg.SupabaseApiTokenType, cfg.SupabaseApiToken)
			req.Header.Set("Authorization", token)
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

func GetTypeByName(cfg *raiden.Config, includedSchema []string, name string) (result objects.Type, err error) {
	MetaLogger.Trace("start fetching type by name from meta")
	sql := sql.GenerateTypeQuery(includedSchema, name) + " limit 1"
	rs, err := ExecuteQuery[[]objects.Type](getBaseUrl(cfg), sql, nil, DefaultInterceptor(cfg), nil)
	if err != nil {
		err = fmt.Errorf("get type error : %s", err)
		return
	}

	if len(rs) == 0 {
		err = fmt.Errorf("get type %s is not found", name)
		return
	}
	MetaLogger.Trace("finish fetching type by name from meta")
	return rs[0], nil
}

func CreateType(cfg *raiden.Config, t objects.Type) (objects.Type, error) {
	MetaLogger.Trace("start create type", "name", t.Name)
	// Execute SQL Query
	sql, err := query.BuildTypeQuery(query.TypeActionCreate, &t)
	if err != nil {
		return objects.Type{}, nil
	}

	_, err = ExecuteQuery[any](getBaseUrl(cfg), sql, nil, DefaultInterceptor(cfg), nil)
	if err != nil {
		return objects.Type{}, fmt.Errorf("create new type %s error : %s", t.Name, err)
	}

	MetaLogger.Trace("finish create type", "name", t.Name)
	return GetTypeByName(cfg, []string{t.Schema}, t.Name)
}

func DeleteType(cfg *raiden.Config, t objects.Type) error {
	MetaLogger.Trace("start delete type", "name", t.Name)
	sql, err := query.BuildTypeQuery(query.TypeActionDelete, &t)
	if err != nil {
		return err
	}

	_, err = ExecuteQuery[any](getBaseUrl(cfg), sql, nil, DefaultInterceptor(cfg), nil)
	if err != nil {
		return fmt.Errorf("delete Type %s error : %s", t.Name, err)
	}

	MetaLogger.Trace("Delete - delete type", "name", t.Name)
	return nil
}

func UpdateType(cfg *raiden.Config, t objects.Type) error {
	MetaLogger.Trace("start update type", "name", t.Name)
	updateSql, err := query.BuildTypeQuery(query.TypeActionUpdate, &t)
	if err != nil {
		return err
	}
	_, err = ExecuteQuery[any](getBaseUrl(cfg), updateSql, nil, DefaultInterceptor(cfg), nil)
	if err != nil {
		return fmt.Errorf("update type %s error : %s", t.Name, err)
	}
	MetaLogger.Trace("finish update type", "name", t.Name)
	return nil
}
