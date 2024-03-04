package cloud

import (
	"fmt"
	"strings"

	"github.com/lib/pq"
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/supabase/drivers/cloud/query"
	"github.com/sev-2/raiden/pkg/supabase/objects"
)

func GetPolicies(cfg *raiden.Config) ([]objects.Policy, error) {
	rs, err := ExecuteQuery[[]objects.Policy](cfg.SupabaseApiUrl, cfg.ProjectId, query.GetPoliciesQuery, DefaultAuthInterceptor(cfg.AccessToken), nil)
	if err != nil {
		err = fmt.Errorf("get policies error : %s", err)
	}
	return rs, err
}

func GetPolicyByName(cfg *raiden.Config, name string) (result objects.Policy, err error) {
	qTemplate := query.GetPoliciesQuery + " where pol.polname = %s limit 1"
	sql := fmt.Sprintf(qTemplate, pq.QuoteLiteral(strings.ToLower(name)))

	// logger.Debug("Get Policy by name - execute : ", sql)
	rs, err := ExecuteQuery[[]objects.Policy](cfg.SupabaseApiUrl, cfg.ProjectId, sql, DefaultAuthInterceptor(cfg.AccessToken), nil)
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
	name := fmt.Sprintf("%q", strings.ToLower(policy.Name))

	definitionClause := ""
	if policy.Definition != "" {
		definitionClause = "USING (" + policy.Definition + ")"
	}
	checkClause := ""
	if policy.Check != nil && *policy.Check != "" {
		checkClause = "WITH CHECK (" + *policy.Check + ")"
	}

	roleList := ""
	for i, role := range policy.Roles {
		if i > 0 {
			roleList += ", "
		}
		roleList += pq.QuoteIdentifier(role)
	}
	sql := fmt.Sprintf(`
	CREATE POLICY %s ON %s.%s
	AS %s
	FOR %s
	TO %s
	%s %s;`, name, policy.Schema, policy.Table, policy.Action, policy.Command, roleList, definitionClause, checkClause)

	// Execute SQL Query
	logger.Debug("Create Policy - execute : ", sql)
	_, err := ExecuteQuery[any](cfg.SupabaseApiUrl, cfg.ProjectId, sql, DefaultAuthInterceptor(cfg.AccessToken), nil)
	if err != nil {
		return objects.Policy{}, fmt.Errorf("create new policy %s error : %s", policy.Name, err)
	}

	return GetPolicyByName(cfg, policy.Name)
}

func UpdatePolicy(cfg *raiden.Config, policy objects.Policy, updatePolicyParams objects.UpdatePolicyParam) error {
	alter := fmt.Sprintf("ALTER POLICY %q ON %s.%s", updatePolicyParams.Name, policy.Schema, policy.Table)

	var nameSql, definitionSql, checkSql, rolesSql string
	for _, ut := range updatePolicyParams.ChangeItems {
		switch ut {
		case objects.UpdatePolicyName:
			if policy.Name != "" {
				nameSql = fmt.Sprintf("%s RENAME TO %s;", alter, policy.Name)
			}
		case objects.UpdatePolicyCheck:
			if policy.Check == nil || (policy.Check != nil && *policy.Check != "") {
				defaultCheck := "true"
				policy.Check = &defaultCheck
			}
			checkSql = fmt.Sprintf("%s WITH CHECK (%s);", alter, *policy.Check)
		case objects.UpdatePolicyDefinition:
			if policy.Definition != "" {
				definitionSql = fmt.Sprintf("%s USING (%s);", alter, policy.Definition)
			}
		case objects.UpdatePolicyRoles:
			if len(policy.Roles) > 0 {
				rolesSql = fmt.Sprintf("%s TO %s;", alter, strings.Join(policy.Roles, ","))
			}
		}
	}

	sql := fmt.Sprintf("BEGIN; %s %s %s %s COMMIT;", definitionSql, checkSql, rolesSql, nameSql)
	// Execute SQL Query
	logger.Debug("Update Policy - execute : ", sql)
	_, err := ExecuteQuery[any](cfg.SupabaseApiUrl, cfg.ProjectId, sql, DefaultAuthInterceptor(cfg.AccessToken), nil)
	if err != nil {
		return fmt.Errorf("update policy %s error : %s", policy.Name, err)
	}
	return nil
}

func DeletePolicy(cfg *raiden.Config, policy objects.Policy) error {
	sql := fmt.Sprintf("DROP POLICY %q ON %s.%s;", policy.Name, policy.Schema, policy.Table)
	// execute delete
	logger.Debug("Delete Policy - execute : ", sql)
	_, err := ExecuteQuery[any](cfg.SupabaseApiUrl, cfg.ProjectId, sql, DefaultAuthInterceptor(cfg.AccessToken), nil)
	if err != nil {
		return fmt.Errorf("delete role %s error : %s", policy.Name, err)
	}
	return nil
}
