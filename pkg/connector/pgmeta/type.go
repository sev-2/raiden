package pgmeta

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
	url := fmt.Sprintf("%s/types", cfg.PgMetaUrl)

	reqInterceptor := func(req *http.Request) error {
		if len(includedSchemas) > 0 {
			req.URL.Query().Set("included_schemas", strings.Join(includedSchemas, ","))
		}

		if len(cfg.JwtToken) > 0 {
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.JwtToken))
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
	rs, err := ExecuteQuery[[]objects.Type](cfg.PgMetaUrl, sql, nil, DefaultAuthInterceptor(cfg.JwtToken), nil)
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
	sql, _ := query.BuildTypeQuery(query.TypeActionCreate, &t)
	_, err := ExecuteQuery[any](cfg.PgMetaUrl, sql, nil, DefaultAuthInterceptor(cfg.JwtToken), nil)
	if err != nil {
		return objects.Type{}, fmt.Errorf("create new type %s error : %s", t.Name, err)
	}

	MetaLogger.Trace("finish create type", "name", t.Name)
	return GetTypeByName(cfg, []string{t.Schema}, t.Name)
}

func DeleteType(cfg *raiden.Config, t objects.Type) error {
	MetaLogger.Trace("start delete type", "name", t.Name)
	sql, _ := query.BuildTypeQuery(query.TypeActionDelete, &t)

	_, err := ExecuteQuery[any](cfg.PgMetaUrl, sql, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("delete Type %s error : %s", t.Name, err)
	}

	MetaLogger.Trace("Delete - delete type", "name", t.Name)
	return nil
}

func UpdateType(cfg *raiden.Config, t objects.Type) error {
	MetaLogger.Trace("start update type", "name", t.Name)
	updateSql, _ := query.BuildTypeQuery(query.TypeActionUpdate, &t)
	_, err := ExecuteQuery[any](cfg.PgMetaUrl, updateSql, nil, DefaultAuthInterceptor(cfg.JwtToken), nil)
	if err != nil {
		return fmt.Errorf("update type %s error : %s", t.Name, err)
	}
	MetaLogger.Trace("finish update type", "name", t.Name)
	return nil
}
