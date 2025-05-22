package pgmeta

import (
	"fmt"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/client/net"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/supabase/query"
	"github.com/sev-2/raiden/pkg/supabase/query/sql"
)

func GetFunctions(cfg *raiden.Config) ([]objects.Function, error) {
	MetaLogger.Trace("start fetching functions from pg meta")

	url := fmt.Sprintf("%s/functions", cfg.PgMetaUrl)
	rs, err := net.Get[[]objects.Function](url, net.DefaultTimeout, DefaultAuthInterceptor(cfg.JwtToken), nil)
	if err != nil {
		err = fmt.Errorf("get functions error : %s", err)
	}

	MetaLogger.Trace("finish fetching functions from pg meta")
	return rs, err
}

func GetFunctionByName(cfg *raiden.Config, schema, name string) (result objects.Function, err error) {
	MetaLogger.Trace("start fetching function by name from meta")
	sql := sql.GenerateFunctionByNameQuery(schema, name) + " limit 1"
	rs, err := ExecuteQuery[[]objects.Function](cfg.PgMetaUrl, sql, nil, DefaultAuthInterceptor(cfg.JwtToken), nil)
	if err != nil {
		err = fmt.Errorf("get function error : %s", err)
		return
	}

	if len(rs) == 0 {
		err = fmt.Errorf("get function %s is not found", name)
		return
	}
	MetaLogger.Trace("finish fetching function by name from meta")
	return rs[0], nil
}

func CreateFunction(cfg *raiden.Config, fn objects.Function) (objects.Function, error) {
	MetaLogger.Trace("start create function", "name", fn.Name)
	// Execute SQL Query
	sql, err := query.BuildFunctionQuery(query.FunctionActionCreate, &fn)
	if err != nil {
		return objects.Function{}, nil
	}

	_, err = ExecuteQuery[any](cfg.PgMetaUrl, sql, nil, DefaultAuthInterceptor(cfg.JwtToken), nil)
	if err != nil {
		return objects.Function{}, fmt.Errorf("create new function %s error : %s", fn.Name, err)
	}

	MetaLogger.Trace("finish create function", "name", fn.Name)
	return GetFunctionByName(cfg, fn.Schema, fn.Name)
}

func DeleteFunction(cfg *raiden.Config, fn objects.Function) error {
	MetaLogger.Trace("start delete function", "name", fn.Name)
	sql, err := query.BuildFunctionQuery(query.FunctionActionDelete, &fn)
	if err != nil {
		return err
	}

	_, err = ExecuteQuery[any](cfg.PgMetaUrl, sql, nil, DefaultAuthInterceptor(cfg.JwtToken), nil)
	if err != nil {
		return fmt.Errorf("delete Function %s error : %s", fn.Name, err)
	}

	MetaLogger.Trace("Delete - delete function", "name", fn.Name)
	return nil
}

func UpdateFunction(cfg *raiden.Config, fn objects.Function) error {
	MetaLogger.Trace("start update function", "name", fn.Name)
	updateSql, err := query.BuildFunctionQuery(query.FunctionActionUpdate, &fn)
	if err != nil {
		return err
	}
	_, err = ExecuteQuery[any](cfg.PgMetaUrl, updateSql, nil, DefaultAuthInterceptor(cfg.JwtToken), nil)
	if err != nil {
		return fmt.Errorf("update function %s error : %s", fn.Name, err)
	}
	MetaLogger.Trace("finish update function", "name", fn.Name)
	return nil
}
