package cloud

import (
	"fmt"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/supabase/query"
	"github.com/sev-2/raiden/pkg/supabase/query/sql"
)

func GetFunctions(cfg *raiden.Config) ([]objects.Function, error) {
	rs, err := ExecuteQuery[[]objects.Function](
		cfg.SupabaseApiUrl, cfg.ProjectId, sql.GenerateFunctionsQuery([]string{"public"}),
		DefaultAuthInterceptor(cfg.AccessToken), nil,
	)
	if err != nil {
		err = fmt.Errorf("get functions error : %s", err)
	}
	return rs, err
}

func GetFunctionByName(cfg *raiden.Config, schema, name string) (result objects.Function, err error) {
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

	return rs[0], nil
}

func CreateFunction(cfg *raiden.Config, fn objects.Function) (objects.Function, error) {
	// Execute SQL Query
	sql, err := query.BuildFunctionQuery(query.FunctionActionCreate, &fn)
	if err != nil {
		return objects.Function{}, nil
	}

	logger.Debug("Create Function - execute : ", sql)
	_, err = ExecuteQuery[any](cfg.SupabaseApiUrl, cfg.ProjectId, sql, DefaultAuthInterceptor(cfg.AccessToken), nil)
	if err != nil {
		return objects.Function{}, fmt.Errorf("create new function %s error : %s", fn.Name, err)
	}

	return GetFunctionByName(cfg, fn.Schema, fn.Name)
}

func DeleteFunction(cfg *raiden.Config, fn objects.Function) error {
	logger.Debug("Create Function - execute : ", fn.CompleteStatement)
	sql, err := query.BuildFunctionQuery(query.FunctionActionDelete, &fn)
	if err != nil {
		return err
	}

	logger.Debug("Delete Function - execute : ", sql)
	_, err = ExecuteQuery[any](cfg.SupabaseApiUrl, cfg.ProjectId, sql, DefaultAuthInterceptor(cfg.AccessToken), nil)
	if err != nil {
		return fmt.Errorf("delete Function %s error : %s", fn.Name, err)
	}
	return nil
}

func UpdateFunction(cfg *raiden.Config, fn objects.Function) error {
	updateSql, err := query.BuildFunctionQuery(query.FunctionActionUpdate, &fn)
	if err != nil {
		return err
	}
	_, err = ExecuteQuery[any](cfg.SupabaseApiUrl, cfg.ProjectId, updateSql, DefaultAuthInterceptor(cfg.AccessToken), nil)
	if err != nil {
		return fmt.Errorf("update function %s error : %s", fn.Name, err)
	}

	return nil
}
