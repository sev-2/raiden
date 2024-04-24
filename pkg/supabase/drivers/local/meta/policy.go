package meta

import (
	"fmt"
	"strings"

	"github.com/lib/pq"
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/supabase/client/net"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/supabase/query"
	"github.com/sev-2/raiden/pkg/supabase/query/sql"
)

func GetPolicies(cfg *raiden.Config) ([]objects.Policy, error) {
	MetaLogger.Trace("Start - fetching policies from meta")
	url := fmt.Sprintf("%s%s/policies", cfg.SupabaseApiUrl, cfg.SupabaseApiBasePath)
	rs, err := net.Get[[]objects.Policy](url, net.DefaultTimeout, nil, nil)
	if err != nil {
		err = fmt.Errorf("get roles error : %s", err)
	}
	MetaLogger.Trace("Finish - fetching policies from meta")
	return rs, err
}

func GetPolicyByName(cfg *raiden.Config, name string) (result objects.Policy, err error) {
	MetaLogger.Trace("Start - fetching policy by name from meta")
	qTemplate := sql.GetPoliciesQuery + " where pol.polname = %s limit 1"
	sql := fmt.Sprintf(qTemplate, pq.QuoteLiteral(strings.ToLower(name)))

	// logger.Trace("Get Policy by name - execute : ", sql)
	rs, err := ExecuteQuery[[]objects.Policy](getBaseUrl(cfg), sql, nil, nil, nil)
	if err != nil {
		err = fmt.Errorf("get policy error : %s", err)
		return
	}

	if len(rs) == 0 {
		err = fmt.Errorf("get policy %s is not found", name)
		return
	}
	MetaLogger.Trace("Finish - fetching policy by name from meta")
	return rs[0], nil
}

func CreatePolicy(cfg *raiden.Config, policy objects.Policy) (objects.Policy, error) {
	MetaLogger.Trace("Start - create policy", "name", policy.Name)
	sql := query.BuildCreatePolicyQuery(policy)

	// Execute SQL Query
	_, err := ExecuteQuery[any](getBaseUrl(cfg), sql, nil, nil, nil)
	if err != nil {
		return objects.Policy{}, fmt.Errorf("create new policy %s error : %s", policy.Name, err)
	}
	MetaLogger.Trace("Finish - create policy", "name", policy.Name)
	return GetPolicyByName(cfg, policy.Name)
}

func UpdatePolicy(cfg *raiden.Config, policy objects.Policy, updatePolicyParams objects.UpdatePolicyParam) error {
	MetaLogger.Trace("Start - update policy", "name", policy.Name)
	sql := query.BuildUpdatePolicyQuery(policy, updatePolicyParams)
	// Execute SQL Query
	_, err := ExecuteQuery[any](getBaseUrl(cfg), sql, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("update policy %s error : %s", policy.Name, err)
	}
	MetaLogger.Trace("Finish - create policy", "name", policy.Name)
	return nil
}

func DeletePolicy(cfg *raiden.Config, policy objects.Policy) error {
	MetaLogger.Trace("Start - delete policy", "name", policy.Name)
	sql := query.BuildDeletePolicyQuery(policy)

	// execute delete
	_, err := ExecuteQuery[any](getBaseUrl(cfg), sql, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("delete role %s error : %s", policy.Name, err)
	}
	MetaLogger.Trace("Finish - create policy", "name", policy.Name)
	return nil
}
