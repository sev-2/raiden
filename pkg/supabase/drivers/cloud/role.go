package cloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/lib/pq"
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/supabase/query"
	"github.com/sev-2/raiden/pkg/supabase/query/sql"
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

var getRoleResponseInterceptor = func(resp *http.Response) error {
	var arrResponse []interface{}
	if err := json.NewDecoder(resp.Body).Decode(&arrResponse); err != nil {
		return err
	}

	decoratedRes := roleResultDecoratorFn(arrResponse)
	byteData, err := json.Marshal(decoratedRes)
	if err != nil {
		return err
	}

	// You can modify the response here by creating a new response with the modified body
	// Note: This is a simplified example and may need to be adapted based on your specific use case
	newResp := *resp
	newResp.Body = io.NopCloser(bytes.NewReader(byteData))
	newResp.ContentLength = int64(len(byteData))

	// Update the original response with the modified response
	*resp = newResp

	return nil
}

func GetRoles(cfg *raiden.Config) ([]objects.Role, error) {
	CloudLogger.Trace("start fetching role from supabase")
	rs, err := ExecuteQuery[[]objects.Role](cfg.SupabaseApiUrl, cfg.ProjectId, sql.GetRolesQuery, DefaultAuthInterceptor(cfg.AccessToken), getRoleResponseInterceptor)
	if err != nil {
		err = fmt.Errorf("get role error : %s", err)
	}
	CloudLogger.Trace("finish fetching role from supabase")
	return rs, err
}

func GetRoleByName(cfg *raiden.Config, name string) (result objects.Role, err error) {
	CloudLogger.Trace("start fetch get singe role by name")
	qTemplate := sql.GetRolesQuery + " where rolname = %s limit 1"
	q := fmt.Sprintf(qTemplate, pq.QuoteLiteral(name))

	rs, err := ExecuteQuery[[]objects.Role](cfg.SupabaseApiUrl, cfg.ProjectId, q, DefaultAuthInterceptor(cfg.AccessToken), getRoleResponseInterceptor)
	if err != nil {
		err = fmt.Errorf("get role error : %s", err)
		return
	}

	if len(rs) == 0 {
		err = fmt.Errorf("get role %s is not found", name)
		return
	}
	CloudLogger.Trace("finish fetch get singe role by name")
	return rs[0], nil
}

func CreateRole(cfg *raiden.Config, role objects.Role) (objects.Role, error) {
	CloudLogger.Trace("start create role", "name", role.Name)
	sql := query.BuildCreateRoleQuery(role)
	// Execute SQL Query
	_, err := ExecuteQuery[any](cfg.SupabaseApiUrl, cfg.ProjectId, sql, DefaultAuthInterceptor(cfg.AccessToken), nil)
	if err != nil {
		return objects.Role{}, fmt.Errorf("create new role %s error : %s", role.Name, err)
	}
	CloudLogger.Trace("finish create role", "name", role.Name)
	return GetRoleByName(cfg, role.Name)
}

func UpdateRole(cfg *raiden.Config, newRole objects.Role, updateRoleParam objects.UpdateRoleParam) error {
	CloudLogger.Trace("start update role", "name", newRole.Name)
	sql := query.BuildUpdateRoleQuery(newRole, updateRoleParam)
	_, err := ExecuteQuery[any](cfg.SupabaseApiUrl, cfg.ProjectId, sql, DefaultAuthInterceptor(cfg.AccessToken), nil)
	if err != nil {
		return fmt.Errorf("update new role %s error : %s", updateRoleParam.OldData.Name, err)
	}
	CloudLogger.Trace("finish update role", "name", newRole.Name)
	return nil
}

func DeleteRole(cfg *raiden.Config, role objects.Role) error {
	CloudLogger.Trace("start delete role", "name", role.Name)
	sql := query.BuildDeleteRoleQuery(role)

	// execute delete
	_, err := ExecuteQuery[any](cfg.SupabaseApiUrl, cfg.ProjectId, sql, DefaultAuthInterceptor(cfg.AccessToken), nil)
	if err != nil {
		return fmt.Errorf("delete role %s error : %s", role.Name, err)
	}
	CloudLogger.Trace("finish delete role", "name", role.Name)
	return nil
}
