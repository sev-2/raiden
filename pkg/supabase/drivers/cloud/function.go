package cloud

import (
	"fmt"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/supabase/query"
	"github.com/sev-2/raiden/pkg/supabase/query/sql"
)

func GetFunctions(cfg *raiden.Config) ([]objects.Function, error) {
	CloudLogger.Trace("start fetching function from supabase")
	q := sql.GenerateFunctionsQuery([]string{"public"})
	rs, err := ExecuteQuery[[]objects.Function](
		cfg.SupabaseApiUrl, cfg.ProjectId, q,
		DefaultAuthInterceptor(cfg.AccessToken), nil,
	)
	if err != nil {
		err = fmt.Errorf("get functions error : %s", err)
	}
	CloudLogger.Trace("finish fetching function from supabase")
	return rs, err
}

func GetFunctionByName(cfg *raiden.Config, schema, name string) (result objects.Function, err error) {
	CloudLogger.Trace("start fetching single function by name")
	sql := sql.GenerateFunctionByNameQuery(schema, name) + " limit 1"
	rs, err := ExecuteQuery[[]objects.Function](cfg.SupabaseApiUrl, cfg.ProjectId, sql, DefaultAuthInterceptor(cfg.AccessToken), nil)
	if err != nil {
		err = fmt.Errorf("get function error : %s", err)
		return
	}

	if len(rs) == 0 {
		err = fmt.Errorf("get function %s is not found", name)
		return
	}
	CloudLogger.Trace("finish fetching single function by name")
	return rs[0], nil
}

func CreateFunction(cfg *raiden.Config, fn objects.Function) (objects.Function, error) {
	CloudLogger.Trace("start create function", "function", fn.Name)
	// Execute SQL Query
	sql, err := query.BuildFunctionQuery(query.FunctionActionCreate, &fn)
	if err != nil {
		return objects.Function{}, nil
	}

	_, err = ExecuteQuery[any](cfg.SupabaseApiUrl, cfg.ProjectId, sql, DefaultAuthInterceptor(cfg.AccessToken), nil)
	if err != nil {
		return objects.Function{}, fmt.Errorf("create new function %s error : %s", fn.Name, err)
	}

	CloudLogger.Trace("finish create function", "function", fn.Name)
	return GetFunctionByName(cfg, fn.Schema, fn.Name)
}

func DeleteFunction(cfg *raiden.Config, fn objects.Function) error {
	CloudLogger.Trace("start delete function", "function", fn.Name)
	sql, err := query.BuildFunctionQuery(query.FunctionActionDelete, &fn)
	if err != nil {
		return err
	}

	_, err = ExecuteQuery[any](cfg.SupabaseApiUrl, cfg.ProjectId, sql, DefaultAuthInterceptor(cfg.AccessToken), nil)
	if err != nil {
		return fmt.Errorf("delete Function %s error : %s", fn.Name, err)
	}
	CloudLogger.Trace("finish delete function", "function", fn.Name)
	return nil
}

func UpdateFunction(cfg *raiden.Config, fn objects.Function) error {
	CloudLogger.Trace("start update function", "function", fn.Name)
	updateSql, err := query.BuildFunctionQuery(query.FunctionActionUpdate, &fn)
	if err != nil {
		return err
	}
	_, err = ExecuteQuery[any](cfg.SupabaseApiUrl, cfg.ProjectId, updateSql, DefaultAuthInterceptor(cfg.AccessToken), nil)
	if err != nil {
		return fmt.Errorf("update function %s error : %s", fn.Name, err)
	}
	CloudLogger.Trace("finish update function", "function", fn.Name)
	return nil
}
