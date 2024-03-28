package cloud

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lib/pq"
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/supabase/query"
	"github.com/sev-2/raiden/pkg/supabase/query/sql"
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

	rs, err := ExecuteQuery[[]objects.Role](cfg.SupabaseApiUrl, cfg.ProjectId, sql.GetRolesQuery, reqInterceptor, resInterceptor)
	if err != nil {
		err = fmt.Errorf("get role error : %s", err)
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

	qTemplate := sql.GetRolesQuery + " where rolname = %s limit 1"
	q := fmt.Sprintf(qTemplate, pq.QuoteLiteral(name))

	rs, err := ExecuteQuery[[]objects.Role](cfg.SupabaseApiUrl, cfg.ProjectId, q, DefaultAuthInterceptor(cfg.AccessToken), resInterceptor)
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
	_, err := ExecuteQuery[any](cfg.SupabaseApiUrl, cfg.ProjectId, sql, DefaultAuthInterceptor(cfg.AccessToken), nil)
	if err != nil {
		return objects.Role{}, fmt.Errorf("create new role %s error : %s", role.Name, err)
	}

	return GetRoleByName(cfg, role.Name)
}

func UpdateRole(cfg *raiden.Config, newRole objects.Role, updateRoleParam objects.UpdateRoleParam) error {
	sql := query.BuildUpdateRoleQuery(newRole, updateRoleParam)
	logger.Debug("Create Role - execute : ", sql)
	_, err := ExecuteQuery[any](cfg.SupabaseApiUrl, cfg.ProjectId, sql, DefaultAuthInterceptor(cfg.AccessToken), nil)
	if err != nil {
		return fmt.Errorf("update new role %s error : %s", updateRoleParam.OldData.Name, err)
	}

	return nil
}

func DeleteRole(cfg *raiden.Config, role objects.Role) error {
	sql := query.BuildDeleteRoleQuery(role)

	// execute delete
	logger.Debug("Delete Role - execute : ", sql)
	_, err := ExecuteQuery[any](cfg.SupabaseApiUrl, cfg.ProjectId, sql, DefaultAuthInterceptor(cfg.AccessToken), nil)
	if err != nil {
		return fmt.Errorf("delete role %s error : %s", role.Name, err)
	}

	return nil
}
