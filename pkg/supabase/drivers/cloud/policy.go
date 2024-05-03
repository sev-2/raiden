package cloud

import (
	"fmt"
	"strings"

	"github.com/lib/pq"
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/supabase/query"
	"github.com/sev-2/raiden/pkg/supabase/query/sql"
)

func GetPolicies(cfg *raiden.Config) ([]objects.Policy, error) {
	CloudLogger.Trace("start fetching policies from supabase")
	rs, err := ExecuteQuery[[]objects.Policy](cfg.SupabaseApiUrl, cfg.ProjectId, sql.GetPoliciesQuery, DefaultAuthInterceptor(cfg.AccessToken), nil)
	if err != nil {
		err = fmt.Errorf("get policies error : %s", err)
	}
	CloudLogger.Trace("finish fetching policies from supabase")
	return rs, err
}

func GetPolicyByName(cfg *raiden.Config, name string) (result objects.Policy, err error) {
	CloudLogger.Trace("start fetching single policy from supabase")
	qTemplate := sql.GetPoliciesQuery + " where pol.polname = %s limit 1"
	sql := fmt.Sprintf(qTemplate, pq.QuoteLiteral(strings.ToLower(name)))

	// logger.Debug("Get Policy by name - execute : ", sql)
	rs, err := ExecuteQuery[[]objects.Policy](cfg.SupabaseApiUrl, cfg.ProjectId, sql, DefaultAuthInterceptor(cfg.AccessToken), nil)
	if err != nil {
		err = fmt.Errorf("get policies error : %s", err)
		return
	}

	if len(rs) == 0 {
		err = fmt.Errorf("get policies %s is not found", name)
		return
	}
	CloudLogger.Trace("finish fetching single policy from supabase")
	return rs[0], nil
}

func CreatePolicy(cfg *raiden.Config, policy objects.Policy) (objects.Policy, error) {
	CloudLogger.Trace("start create policy", "name", policy.Name)
	sql := query.BuildCreatePolicyQuery(policy)

	// Execute SQL Query
	_, err := ExecuteQuery[any](cfg.SupabaseApiUrl, cfg.ProjectId, sql, DefaultAuthInterceptor(cfg.AccessToken), nil)
	if err != nil {
		return objects.Policy{}, fmt.Errorf("create new policy %s error : %s", policy.Name, err)
	}

	CloudLogger.Trace("finish create policy", "name", policy.Name)
	return GetPolicyByName(cfg, policy.Name)
}

func UpdatePolicy(cfg *raiden.Config, policy objects.Policy, updatePolicyParams objects.UpdatePolicyParam) error {
	CloudLogger.Trace("start update policy", "name", policy.Name)
	sql := query.BuildUpdatePolicyQuery(policy, updatePolicyParams)
	// Execute SQL Query
	_, err := ExecuteQuery[any](cfg.SupabaseApiUrl, cfg.ProjectId, sql, DefaultAuthInterceptor(cfg.AccessToken), nil)
	if err != nil {
		return fmt.Errorf("update policy %s error : %s", policy.Name, err)
	}
	CloudLogger.Trace("finish update policy", "name", policy.Name)
	return nil
}

func DeletePolicy(cfg *raiden.Config, policy objects.Policy) error {
	CloudLogger.Trace("start delete policy", "name", policy.Name)
	sql := query.BuildDeletePolicyQuery(policy)

	_, err := ExecuteQuery[any](cfg.SupabaseApiUrl, cfg.ProjectId, sql, DefaultAuthInterceptor(cfg.AccessToken), nil)
	if err != nil {
		return fmt.Errorf("delete role %s error : %s", policy.Name, err)
	}
	CloudLogger.Trace("finish delete policy", "name", policy.Name)
	return nil
}
