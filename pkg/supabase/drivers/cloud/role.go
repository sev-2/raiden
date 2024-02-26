package cloud

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lib/pq"
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/supabase/drivers/cloud/query"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/valyala/fasthttp"
)

func roleFindConfigFn(role any) []any {
	if roleMap, isMapAny := role.(map[string]any); isMapAny {
		if configValue, exist := roleMap["config"]; exist {
			if configArr, isArrayAny := configValue.([]any); isArrayAny {
				return configArr
			}
		}
	}
	return nil
}

func roleConfigsToMapFn(configs []any) map[string]any {
	mapConfig := make(map[string]any)
	for _, configItem := range configs {
		if configItemStr, isString := configItem.(string); isString {
			configItemSplitted := strings.Split(configItemStr, "=")
			if len(configItemSplitted) == 2 {
				mapConfig[configItemSplitted[0]] = configItemSplitted[1]
			}
		}
	}
	return mapConfig
}

func roleResultDecoratorFn(result any) any {
	if roles, isRolesArr := result.([]any); isRolesArr {
		for roleIndex := range roles {
			roleItem := roles[roleIndex]
			if foundConfig := roleFindConfigFn(roleItem); foundConfig != nil {
				config := roleConfigsToMapFn(foundConfig)
				if config != nil {
					roleItem.(map[string]any)["config"] = config
				}
			}
		}
	}
	return result
}

func GetRoles(cfg *raiden.Config) ([]objects.Role, error) {
	reqInterceptor := func(req *fasthttp.Request) error {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.AccessToken))
		return nil
	}

	resInterceptor := func(res *fasthttp.Response) error {
		var arrResponse []any
		if err := json.Unmarshal(res.Body(), &arrResponse); err != nil {
			return err
		}

		decoratedRes := roleResultDecoratorFn(arrResponse)
		byteData, err := json.Marshal(decoratedRes)
		if err != nil {
			return err
		}

		res.SetBodyRaw(byteData)
		return nil
	}

	rs, err := ExecuteQuery[[]objects.Role](cfg.SupabaseApiUrl, cfg.ProjectId, query.GetRolesQuery, reqInterceptor, resInterceptor)
	if err != nil {
		err = fmt.Errorf("get roles error : %s", err)
	}

	return rs, err
}

func GetRoleByName(cfg *raiden.Config, name string) (result objects.Role, err error) {
	resInterceptor := func(res *fasthttp.Response) error {
		var arrResponse []any
		if err := json.Unmarshal(res.Body(), &arrResponse); err != nil {
			return err
		}

		decoratedRes := roleResultDecoratorFn(arrResponse)
		byteData, err := json.Marshal(decoratedRes)
		if err != nil {
			return err
		}

		res.SetBodyRaw(byteData)
		return nil
	}

	qTemplate := query.GetRolesQuery + " where rolname = %s limit 1"
	q := fmt.Sprintf(qTemplate, pq.QuoteLiteral(name))

	rs, err := ExecuteQuery[[]objects.Role](cfg.SupabaseApiUrl, cfg.ProjectId, q, DefaultAuthInterceptor(cfg.AccessToken), resInterceptor)
	if err != nil {
		err = fmt.Errorf("get roles error : %s", err)
		return
	}

	if len(rs) == 0 {
		err = fmt.Errorf("get role %s is not found", name)
		return
	}

	return rs[0], nil
}

func CreateRole(cfg *raiden.Config, role objects.Role) (objects.Role, error) {
	var createRolClauses []string

	canCreateDBClause := "NOCREATEDB"
	if role.CanCreateDB {
		canCreateDBClause = "CREATEDB"
	}
	createRolClauses = append(createRolClauses, canCreateDBClause)

	canCreateRoleClause := "NOCREATEROLE"
	if role.CanCreateRole {
		canCreateRoleClause = "CREATEROLE"
	}
	createRolClauses = append(createRolClauses, canCreateRoleClause)

	canBypassRLSClause := "NOBYPASSRLS"
	if role.CanBypassRLS {
		canBypassRLSClause = "BYPASSRLS"
	}
	createRolClauses = append(createRolClauses, canBypassRLSClause)

	canLoginClause := "NOLOGIN"
	if role.CanLogin {
		canLoginClause = "LOGIN"
	}
	createRolClauses = append(createRolClauses, canLoginClause)

	inheritRoleClause := "NOINHERIT"
	if role.InheritRole {
		inheritRoleClause = "INHERIT"
	}
	createRolClauses = append(createRolClauses, inheritRoleClause)

	isSuperuserClause := "NOSUPERUSER"
	if role.IsSuperuser {
		isSuperuserClause = "SUPERUSER"
	}
	createRolClauses = append(createRolClauses, isSuperuserClause)

	isReplicationRoleClause := "NOREPLICATION"
	if role.IsReplicationRole {
		isReplicationRoleClause = "REPLICATION"
	}
	createRolClauses = append(createRolClauses, isReplicationRoleClause)

	connectionLimitClause := fmt.Sprintf("CONNECTION LIMIT %d", role.ConnectionLimit)
	createRolClauses = append(createRolClauses, connectionLimitClause)

	// passwordClause := ""
	// if role.Password != "" {
	// 	passwordClause = fmt.Sprintf("PASSWORD %s", literal(role.Password))
	// }

	if role.ValidUntil != nil {
		validUntilClause := fmt.Sprintf("VALID UNTIL '%s'", role.ValidUntil.Format(raiden.DefaultRoleValidUntilLayout))
		createRolClauses = append(createRolClauses, validUntilClause)
	}

	configClause := ""
	if len(role.Config) > 0 {
		var configStrings []string
		for k, v := range role.Config {
			if k != "" && v != "" {
				configStrings = append(configStrings, fmt.Sprintf("ALTER ROLE %s SET %s = %s;", role.Name, k, v))
			}
		}
		configClause = strings.Join(configStrings, " ")
	}

	sql := fmt.Sprintf(`
	BEGIN;
	do $$
	BEGIN
		IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'vote_manager') THEN
			CREATE ROLE %s WITH %s;
		END IF;
	END $$;
	%s
	COMMIT;`,
		role.Name, strings.Join(createRolClauses, "\n"),
		configClause,
	)

	// Execute SQL Query
	logger.Debug("Create Role - execute : ", sql)
	_, err := ExecuteQuery[any](cfg.SupabaseApiUrl, cfg.ProjectId, sql, DefaultAuthInterceptor(cfg.AccessToken), nil)
	if err != nil {
		return objects.Role{}, fmt.Errorf("create new role %s error : %s", role.Name, err)
	}

	return GetRoleByName(cfg, role.Name)
}

func UpdateRole(cfg *raiden.Config, newRole objects.Role, updateRoleParam objects.UpdateRoleParam) error {
	alter := fmt.Sprintf("ALTER ROLE %s ", updateRoleParam.OldData.Name)

	var updateRoleClause, nameClause, configClause string
	var updateRoleClauses []string

	for _, item := range updateRoleParam.ChangeItems {
		switch item {
		case objects.UpdateConnectionLimit:
			updateRoleClauses = append(updateRoleClauses, fmt.Sprintf("CONNECTION LIMIT %d", newRole.ConnectionLimit))
		case objects.UpdateRoleName:
			if updateRoleParam.OldData.Name != newRole.Name {
				nameClause = fmt.Sprintf("%s RENAME TO %s", alter, newRole.Name)
			}
		case objects.UpdateRoleIsReplication:
			if newRole.CanLogin {
				updateRoleClauses = append(updateRoleClauses, "REPLICATION")
			} else {
				updateRoleClauses = append(updateRoleClauses, "NOREPLICATION")
			}
		case objects.UpdateRoleIsSuperUser:
			if newRole.IsSuperuser {
				updateRoleClauses = append(updateRoleClauses, "SUPERUSER")
			} else {
				updateRoleClauses = append(updateRoleClauses, "NOSUPERUSER")
			}
		case objects.UpdateRoleInheritRole:
			if newRole.InheritRole {
				updateRoleClauses = append(updateRoleClauses, "INHERIT")
			} else {
				updateRoleClauses = append(updateRoleClauses, "NOINHERIT")
			}
		case objects.UpdateRoleCanBypassRls:
			if newRole.CanBypassRLS {
				updateRoleClauses = append(updateRoleClauses, "BYPASSRLS")
			} else {
				updateRoleClauses = append(updateRoleClauses, "NOBYPASSRLS")
			}
		case objects.UpdateRoleCanCreateDb:
			if newRole.CanCreateDB {
				updateRoleClauses = append(updateRoleClauses, "CREATEDB")
			} else {
				updateRoleClauses = append(updateRoleClauses, "NOCREATEDB")
			}
		case objects.UpdateRoleCanCreateRole:
			if newRole.CanCreateRole {
				updateRoleClauses = append(updateRoleClauses, "CREATEROLE")
			} else {
				updateRoleClauses = append(updateRoleClauses, "NOCREATEROLE")
			}
		case objects.UpdateRoleCanLogin:
			if newRole.CanLogin {
				updateRoleClauses = append(updateRoleClauses, "LOGIN")
			} else {
				updateRoleClauses = append(updateRoleClauses, "NOLOGIN")
			}
		case objects.UpdateRoleValidUntil:
			validUntilClause := fmt.Sprintf("VALID UNTIL '%s'", newRole.ValidUntil.Format(raiden.DefaultRoleValidUntilLayout))
			updateRoleClauses = append(updateRoleClauses, validUntilClause)
		case objects.UpdateRoleConfig:
			var configStrings []string
			for k, v := range newRole.Config {
				if k != "" && v != "" {
					configStrings = append(configStrings, fmt.Sprintf("%s SET %s = %s;", alter, k, v))
				}
			}
			configClause = strings.Join(configStrings, "\n")
		}
	}

	if len(updateRoleClauses) > 0 {
		updateRoleClause = fmt.Sprintf("%s %s;", alter, strings.Join(updateRoleClauses, "\n"))
	}

	sql := fmt.Sprintf(`
		BEGIN; %s %s %s COMMIT;
	`, updateRoleClause, configClause, nameClause)

	logger.Debug("Create Role - execute : ", sql)
	_, err := ExecuteQuery[any](cfg.SupabaseApiUrl, cfg.ProjectId, sql, DefaultAuthInterceptor(cfg.AccessToken), nil)
	if err != nil {
		return fmt.Errorf("update new role %s error : %s", updateRoleParam.OldData.Name, err)
	}

	return nil
}

func DeleteRole(cfg *raiden.Config, role objects.Role) error {
	sql := fmt.Sprintf("DROP ROLE %s;", role.Name)

	// execute delete
	logger.Debug("Delete Role - execute : ", sql)
	_, err := ExecuteQuery[any](cfg.SupabaseApiUrl, cfg.ProjectId, sql, DefaultAuthInterceptor(cfg.AccessToken), nil)
	if err != nil {
		return fmt.Errorf("delete role %s error : %s", role.Name, err)
	}

	return nil
}
