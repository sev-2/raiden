package meta

import (
	"errors"
	"fmt"
	"sync"

	"github.com/lib/pq"
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/client/net"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/supabase/query"
	"github.com/sev-2/raiden/pkg/supabase/query/sql"
)

func GetRoles(cfg *raiden.Config) ([]objects.Role, error) {
	MetaLogger.Trace("start fetching roles from meta")
	url := fmt.Sprintf("%s%s/roles", cfg.SupabaseApiUrl, cfg.SupabaseApiBasePath)
	rs, err := net.Get[[]objects.Role](url, net.DefaultTimeout, DefaultInterceptor(cfg), nil)
	if err != nil {
		err = fmt.Errorf("get roles error : %s", err)
	}
	MetaLogger.Trace("finish fetching roles from meta")
	return rs, err
}

func GetRoleMemberships(cfg *raiden.Config, includedSchema []string) ([]objects.RoleMembership, error) {
	MetaLogger.Trace("start fetching role memberships from meta")
	rs, err := ExecuteQuery[[]objects.RoleMembership](getBaseUrl(cfg), sql.GenerateGetRoleMembershipsQuery(includedSchema), nil, DefaultInterceptor(cfg), nil)
	if err != nil {
		err = fmt.Errorf("get role membership error : %s", err)
	}
	MetaLogger.Trace("finish fetching role memberships from meta")
	return rs, err
}

func GetRoleByName(cfg *raiden.Config, name string) (result objects.Role, err error) {
	MetaLogger.Trace("start fetching role by name from meta")
	qTemplate := sql.GetRolesQuery + " where rolname = %s limit 1"
	q := fmt.Sprintf(qTemplate, pq.QuoteLiteral(name))

	rs, err := ExecuteQuery[[]objects.Role](getBaseUrl(cfg), q, nil, DefaultInterceptor(cfg), nil)
	if err != nil {
		err = fmt.Errorf("get role error : %s", err)
		return
	}

	if len(rs) == 0 {
		err = fmt.Errorf("get role %s is not found", name)
		return
	}
	MetaLogger.Trace("finish fetching role by name from meta")
	return rs[0], nil
}

func CreateRole(cfg *raiden.Config, role objects.Role) (objects.Role, error) {
	MetaLogger.Trace("start create role", "name", role.Name)
	sql := query.BuildCreateRoleQuery(role)
	// Execute SQL Query
	_, err := ExecuteQuery[any](getBaseUrl(cfg), sql, nil, DefaultInterceptor(cfg), nil)
	if err != nil {
		return objects.Role{}, fmt.Errorf("create new role %s error : %s", role.Name, err)
	}
	MetaLogger.Trace("finish create role", "name", role.Name)
	return GetRoleByName(cfg, role.Name)
}

func UpdateRole(cfg *raiden.Config, newRole objects.Role, updateRoleParam objects.UpdateRoleParam) error {
	MetaLogger.Trace("start update role", "name", newRole.Name)
	roleName := newRole.Name
	if roleName == "" {
		roleName = updateRoleParam.OldData.Name
	}

	if len(updateRoleParam.ChangeItems) == 0 && len(updateRoleParam.ChangeInheritItems) == 0 {
		return fmt.Errorf("update role %s has no changes", roleName)
	}

	var collectedErrors []error

	if len(updateRoleParam.ChangeItems) > 0 {
		sql := query.BuildUpdateRoleQuery(newRole, updateRoleParam)
		if _, err := ExecuteQuery[any](getBaseUrl(cfg), sql, nil, DefaultInterceptor(cfg), nil); err != nil {
			collectedErrors = append(collectedErrors, fmt.Errorf("update role %s error : %s", updateRoleParam.OldData.Name, err))
		}
	}

	if len(updateRoleParam.ChangeInheritItems) > 0 {
		if errs := updateRoleInheritances(cfg, roleName, updateRoleParam.ChangeInheritItems); len(errs) > 0 {
			collectedErrors = append(collectedErrors, errs...)
		}
	}

	if len(collectedErrors) > 0 {
		return errors.Join(collectedErrors...)
	}

	MetaLogger.Trace("finish update role", "name", roleName)
	return nil
}

func DeleteRole(cfg *raiden.Config, role objects.Role) error {
	MetaLogger.Trace("start delete role", "name", role.Name)
	sql := query.BuildDeleteRoleQuery(role)

	// execute delete
	_, err := ExecuteQuery[any](getBaseUrl(cfg), sql, nil, DefaultInterceptor(cfg), nil)
	if err != nil {
		return fmt.Errorf("delete role %s error : %s", role.Name, err)
	}
	MetaLogger.Trace("finish delete role", "name", role.Name)
	return nil
}

func updateRoleInheritances(cfg *raiden.Config, roleName string, items []objects.UpdateRoleInheritItem) []error {
	validItems := make([]objects.UpdateRoleInheritItem, 0, len(items))
	for i := range items {
		it := items[i]
		if it.Role.Name == "" {
			continue
		}
		validItems = append(validItems, it)
	}

	if len(validItems) == 0 {
		return nil
	}

	wg := sync.WaitGroup{}
	errChan := make(chan error)

	for i := range validItems {
		item := validItems[i]
		wg.Add(1)
		go func(w *sync.WaitGroup, inheritItem objects.UpdateRoleInheritItem) {
			defer w.Done()

			sql, err := query.BuildRoleInheritQuery(roleName, inheritItem.Role.Name, inheritItem.Type)
			if err != nil {
				errChan <- err
				return
			}

			_, err = ExecuteQuery[any](getBaseUrl(cfg), sql, nil, DefaultInterceptor(cfg), nil)
			if err != nil {
				action := "grant"
				if inheritItem.Type == objects.UpdateRoleInheritRevoke {
					action = "revoke"
				}
				errChan <- fmt.Errorf("%s role %s for %s error : %s", action, inheritItem.Role.Name, roleName, err)
				return
			}

			errChan <- nil
		}(&wg, item)
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	errs := make([]error, 0)
	for err := range errChan {
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}
