package meta

import (
	"fmt"

	"github.com/lib/pq"
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/supabase/client/net"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/supabase/query"
	"github.com/sev-2/raiden/pkg/supabase/query/sql"
)

func GetRoles(cfg *raiden.Config) ([]objects.Role, error) {
	MetaLogger.Trace("Start - fetching roles from meta")
	url := fmt.Sprintf("%s%s/roles", cfg.SupabaseApiUrl, cfg.SupabaseApiBasePath)
	rs, err := net.Get[[]objects.Role](url, net.DefaultTimeout, nil, nil)
	if err != nil {
		err = fmt.Errorf("get roles error : %s", err)
	}
	MetaLogger.Trace("Finish - fetching roles from meta")
	return rs, err
}

func GetRoleByName(cfg *raiden.Config, name string) (result objects.Role, err error) {
	MetaLogger.Trace("Start - fetching role by name from meta")
	qTemplate := sql.GetRolesQuery + " where rolname = %s limit 1"
	q := fmt.Sprintf(qTemplate, pq.QuoteLiteral(name))

	rs, err := ExecuteQuery[[]objects.Role](getBaseUrl(cfg), q, nil, nil, nil)
	if err != nil {
		err = fmt.Errorf("get role error : %s", err)
		return
	}

	if len(rs) == 0 {
		err = fmt.Errorf("get role %s is not found", name)
		return
	}
	MetaLogger.Trace("Finish - fetching role by name from meta")
	return rs[0], nil
}

func CreateRole(cfg *raiden.Config, role objects.Role) (objects.Role, error) {
	MetaLogger.Trace("Start - create role", "name", role.Name)
	sql := query.BuildCreateRoleQuery(role)
	// Execute SQL Query
	_, err := ExecuteQuery[any](getBaseUrl(cfg), sql, nil, nil, nil)
	if err != nil {
		return objects.Role{}, fmt.Errorf("create new role %s error : %s", role.Name, err)
	}
	MetaLogger.Trace("Finish - create role", "name", role.Name)
	return GetRoleByName(cfg, role.Name)
}

func UpdateRole(cfg *raiden.Config, newRole objects.Role, updateRoleParam objects.UpdateRoleParam) error {
	MetaLogger.Trace("Start - update role", "name", newRole.Name)
	sql := query.BuildUpdateRoleQuery(newRole, updateRoleParam)
	_, err := ExecuteQuery[any](getBaseUrl(cfg), sql, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("update new role %s error : %s", updateRoleParam.OldData.Name, err)
	}
	MetaLogger.Trace("Finish - update role", "name", newRole.Name)
	return nil
}

func DeleteRole(cfg *raiden.Config, role objects.Role) error {
	MetaLogger.Trace("Start - delete role", "name", role.Name)
	sql := query.BuildDeleteRoleQuery(role)

	// execute delete
	_, err := ExecuteQuery[any](getBaseUrl(cfg), sql, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("delete role %s error : %s", role.Name, err)
	}
	MetaLogger.Trace("Finish - delete role", "name", role.Name)
	return nil
}
