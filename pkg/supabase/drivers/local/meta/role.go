package meta

import (
	"fmt"

	"github.com/lib/pq"
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/supabase/client"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/supabase/query"
	"github.com/sev-2/raiden/pkg/supabase/query/sql"
)

func GetRoles(cfg *raiden.Config) ([]objects.Role, error) {
	url := fmt.Sprintf("%s%s/roles", cfg.SupabaseApiUrl, cfg.SupabaseApiBasePath)
	rs, err := client.Get[[]objects.Role](url, client.DefaultTimeout, nil, nil)
	if err != nil {
		err = fmt.Errorf("get roles error : %s", err)
	}
	return rs, err
}

func GetRoleByName(cfg *raiden.Config, name string) (result objects.Role, err error) {
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

	return rs[0], nil
}

func CreateRole(cfg *raiden.Config, role objects.Role) (objects.Role, error) {
	sql := query.BuildCreateRoleQuery(role)
	// Execute SQL Query
	logger.Debug("Create Role - execute : ", sql)
	_, err := ExecuteQuery[any](getBaseUrl(cfg), sql, nil, nil, nil)
	if err != nil {
		return objects.Role{}, fmt.Errorf("create new role %s error : %s", role.Name, err)
	}

	return GetRoleByName(cfg, role.Name)
}

func UpdateRole(cfg *raiden.Config, newRole objects.Role, updateRoleParam objects.UpdateRoleParam) error {
	sql := query.BuildUpdateRoleQuery(newRole, updateRoleParam)
	logger.Debug("Create Role - execute : ", sql)
	_, err := ExecuteQuery[any](getBaseUrl(cfg), sql, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("update new role %s error : %s", updateRoleParam.OldData.Name, err)
	}

	return nil
}

func DeleteRole(cfg *raiden.Config, role objects.Role) error {
	sql := query.BuildDeleteRoleQuery(role)

	// execute delete
	logger.Debug("Delete Role - execute : ", sql)
	_, err := ExecuteQuery[any](getBaseUrl(cfg), sql, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("delete role %s error : %s", role.Name, err)
	}

	return nil
}
