package meta

import (
	"fmt"
	"strings"

	"github.com/lib/pq"
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/supabase/client"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/supabase/query"
	"github.com/sev-2/raiden/pkg/supabase/query/sql"
)

func GetPolicies(cfg *raiden.Config) ([]objects.Policy, error) {
	url := fmt.Sprintf("%s%s/policies", cfg.SupabaseApiUrl, cfg.SupabaseApiBasePath)
	rs, err := client.Get[[]objects.Policy](url, client.DefaultTimeout, nil, nil)
	if err != nil {
		err = fmt.Errorf("get roles error : %s", err)
	}
	return rs, err
}

func GetPolicyByName(cfg *raiden.Config, name string) (result objects.Policy, err error) {
	qTemplate := sql.GetPoliciesQuery + " where pol.polname = %s limit 1"
	sql := fmt.Sprintf(qTemplate, pq.QuoteLiteral(strings.ToLower(name)))

	// logger.Debug("Get Policy by name - execute : ", sql)
	rs, err := ExecuteQuery[[]objects.Policy](getBaseUrl(cfg), sql, nil, nil, nil)
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

func CreatePolicy(cfg *raiden.Config, policy objects.Policy) (objects.Policy, error) {
	sql := query.BuildCreatePolicyQuery(policy)

	// Execute SQL Query
	logger.Debug("Create Policy - execute : ", sql)
	_, err := ExecuteQuery[any](getBaseUrl(cfg), sql, nil, nil, nil)
	if err != nil {
		return objects.Policy{}, fmt.Errorf("create new policy %s error : %s", policy.Name, err)
	}

	return GetPolicyByName(cfg, policy.Name)
}

func UpdatePolicy(cfg *raiden.Config, policy objects.Policy, updatePolicyParams objects.UpdatePolicyParam) error {
	sql := query.BuildUpdatePolicyQuery(policy, updatePolicyParams)
	// Execute SQL Query
	logger.Debug("Update Policy - execute : ", sql)
	_, err := ExecuteQuery[any](getBaseUrl(cfg), sql, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("update policy %s error : %s", policy.Name, err)
	}
	return nil
}

func DeletePolicy(cfg *raiden.Config, policy objects.Policy) error {
	sql := query.BuildDeletePolicyQuery(policy)

	// execute delete
	logger.Debug("Delete Policy - execute : ", sql)
	_, err := ExecuteQuery[any](getBaseUrl(cfg), sql, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("delete role %s error : %s", policy.Name, err)
	}
	return nil
}
