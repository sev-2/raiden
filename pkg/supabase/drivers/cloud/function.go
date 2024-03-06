package cloud

import (
	"fmt"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/supabase/drivers/cloud/query"
	"github.com/sev-2/raiden/pkg/supabase/objects"
)

type FunctionAction string

const (
	FunctionActionCreate FunctionAction = "create"
	FunctionActionUpdate FunctionAction = "update"
	FunctionActionDelete FunctionAction = "delete"
)

func GetFunctions(cfg *raiden.Config) ([]objects.Function, error) {
	rs, err := ExecuteQuery[[]objects.Function](
		cfg.SupabaseApiUrl, cfg.ProjectId, query.GenerateFunctionsQuery([]string{"public"}),
		DefaultAuthInterceptor(cfg.AccessToken), nil,
	)
	if err != nil {
		err = fmt.Errorf("get functions error : %s", err)
	}
	return rs, err
}

func GetFunctionByName(cfg *raiden.Config, schema, name string) (result objects.Function, err error) {
	sql := query.GenerateFunctionByNameQuery(schema, name) + " limit 1"
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
	sql, err := getFunctionQuery(FunctionActionCreate, &fn)
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
	sql, err := getFunctionQuery(FunctionActionDelete, &fn)
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
	deleteSql, err := getFunctionQuery(FunctionActionDelete, &fn)
	if err != nil {
		return err
	}

	createSql, err := getFunctionQuery(FunctionActionCreate, &fn)
	if err != nil {
		return err
	}

	updateSql := fmt.Sprintf(`
		BEGIN; 
			%s 
			%s  
		COMMIT;
	`, deleteSql, createSql)
	_, err = ExecuteQuery[any](cfg.SupabaseApiUrl, cfg.ProjectId, updateSql, DefaultAuthInterceptor(cfg.AccessToken), nil)
	if err != nil {
		return fmt.Errorf("update function %s error : %s", fn.Name, err)
	}

	return nil
}

func getFunctionQuery(action FunctionAction, fn *objects.Function) (string, error) {
	switch action {
	case FunctionActionCreate:
		return fn.CompleteStatement + ";", nil
	case FunctionActionDelete:
		return fmt.Sprintf("DROP FUNCTION %s.%s;", fn.Schema, fn.Name), nil
	default:
		return "", fmt.Errorf("generate function sql with type '%s' is not available", action)
	}
}
