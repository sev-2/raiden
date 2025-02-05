package cloud

import (
	"fmt"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/supabase/query"
	"github.com/sev-2/raiden/pkg/supabase/query/sql"
)

func GetTypes(cfg *raiden.Config, includedSchemas []string) ([]objects.Type, error) {
	CloudLogger.Trace("start fetching types from supabase")
	q := sql.GenerateTypesQuery(includedSchemas)
	rs, err := ExecuteQuery[[]objects.Type](cfg.SupabaseApiUrl, cfg.ProjectId, q, DefaultAuthInterceptor(cfg.AccessToken), nil)
	if err != nil {
		err = fmt.Errorf("get types error : %s", err)
	}
	CloudLogger.Trace("finish fetching types from supabase")
	return rs, err
}

func GetTypeByName(cfg *raiden.Config, includedSchema []string, name string) (result objects.Type, err error) {
	CloudLogger.Trace("start fetching single type by name")
	sql := sql.GenerateTypeQuery(includedSchema, name) + " limit 1"
	rs, err := ExecuteQuery[[]objects.Type](cfg.SupabaseApiUrl, cfg.ProjectId, sql, DefaultAuthInterceptor(cfg.AccessToken), nil)
	if err != nil {
		err = fmt.Errorf("get type error : %s", err)
		return
	}

	if len(rs) == 0 {
		err = fmt.Errorf("get type %s is not found", name)
		return
	}
	CloudLogger.Trace("finish fetching single type by name")
	return rs[0], nil
}

func CreateType(cfg *raiden.Config, fn objects.Type) (objects.Type, error) {
	CloudLogger.Trace("start create type", "type", fn.Name)
	// Execute SQL Query
	sql, err := query.BuildTypeQuery(query.TypeActionCreate, &fn)
	if err != nil {
		return objects.Type{}, nil
	}

	sql = cleanupQueryParam(sql)

	_, err = ExecuteQuery[any](cfg.SupabaseApiUrl, cfg.ProjectId, sql, DefaultAuthInterceptor(cfg.AccessToken), nil)
	if err != nil {
		return objects.Type{}, fmt.Errorf("create new type %s error : %s", fn.Name, err)
	}

	CloudLogger.Trace("finish create type", "type", fn.Name)
	return GetTypeByName(cfg, []string{fn.Schema}, fn.Name)
}

func DeleteType(cfg *raiden.Config, fn objects.Type) error {
	CloudLogger.Trace("start delete type", "type", fn.Name)
	sql, err := query.BuildTypeQuery(query.TypeActionDelete, &fn)
	if err != nil {
		return err
	}

	_, err = ExecuteQuery[any](cfg.SupabaseApiUrl, cfg.ProjectId, sql, DefaultAuthInterceptor(cfg.AccessToken), nil)
	if err != nil {
		return fmt.Errorf("delete Type %s error : %s", fn.Name, err)
	}
	CloudLogger.Trace("finish delete type", "type", fn.Name)
	return nil
}

func UpdateType(cfg *raiden.Config, fn objects.Type) error {
	CloudLogger.Trace("start update type", "type", fn.Name)
	updateSql, err := query.BuildTypeQuery(query.TypeActionUpdate, &fn)
	if err != nil {
		return err
	}
	updateSql = cleanupQueryParam(updateSql)
	_, err = ExecuteQuery[any](cfg.SupabaseApiUrl, cfg.ProjectId, updateSql, DefaultAuthInterceptor(cfg.AccessToken), nil)
	if err != nil {
		return fmt.Errorf("update type %s error : %s", fn.Name, err)
	}
	CloudLogger.Trace("finish update type", "type", fn.Name)
	return nil
}
