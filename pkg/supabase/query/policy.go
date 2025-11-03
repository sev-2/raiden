package query

import (
	"fmt"
	"strings"

	"github.com/lib/pq"
	"github.com/sev-2/raiden/pkg/supabase/objects"
)

func BuildCreatePolicyQuery(policy objects.Policy) string {
	name := fmt.Sprintf("%q", strings.ToLower(policy.Name))

	action := strings.ToUpper(strings.TrimSpace(policy.Action))
	if action == "" {
		action = "PERMISSIVE"
	}
	command := strings.ToUpper(strings.TrimSpace(string(policy.Command)))
	if command == "" {
		command = string(objects.PolicyCommandAll)
	}
	schemaIdent := pq.QuoteIdentifier(policy.Schema)
	tableIdent := pq.QuoteIdentifier(policy.Table)
	tableFQN := fmt.Sprintf("%s.%s", schemaIdent, tableIdent)
	tableFQNText := fmt.Sprintf("%s.%s", policy.Schema, policy.Table)

	definitionClause := ""
	definition := strings.TrimSpace(policy.Definition)
	if command != string(objects.PolicyCommandInsert) && definition != "" {
		definitionClause = "USING (" + definition + ")"
	}
	checkClause := ""
	checkValue := ""
	if policy.Check != nil {
		checkValue = strings.TrimSpace(*policy.Check)
	}
	checkAllowed := command != string(objects.PolicyCommandSelect) && command != string(objects.PolicyCommandDelete)
	if checkAllowed && checkValue != "" {
		checkClause = "WITH CHECK (" + checkValue + ")"
	} else if command == string(objects.PolicyCommandInsert) {
		checkClause = "WITH CHECK (true)"
	}

	roleList := ""
	grantAccessTables := []string{}
	privileges := privilegesForCommand(command)
	for i, role := range policy.Roles {
		if i > 0 {
			roleList += ", "
		}
		roleList += pq.QuoteIdentifier(role)

		for _, privilege := range privileges {
			grantAccessTables = append(grantAccessTables, fmt.Sprintf(`
			IF NOT HAS_TABLE_PRIVILEGE('%s', '%s', '%s') THEN
				GRANT %s ON %s TO %s;
			END IF;
		`, role, tableFQNText, privilege, privilege, tableFQN, pq.QuoteIdentifier(role)))
		}
	}

	if roleList == "" {
		roleList = "PUBLIC"
	}

	createQuery := fmt.Sprintf(`
	CREATE POLICY %s ON %s.%s
	AS %s
	FOR %s
	TO %s
	%s %s;
	`, name, schemaIdent, tableIdent, action, command, roleList, definitionClause, checkClause)

	grantStatements := strings.Join(grantAccessTables, "\n")

	grantAccessQuery := fmt.Sprintf(`
		DO $$
		BEGIN
			%s
			%s
		END $$;
	`, createQuery, grantStatements)

	return grantAccessQuery
}

func BuildUpdatePolicyQuery(policy objects.Policy, updatePolicyParams objects.UpdatePolicyParam) string {
	schemaIdent := pq.QuoteIdentifier(policy.Schema)
	tableIdent := pq.QuoteIdentifier(policy.Table)
	tableFQN := fmt.Sprintf("%s.%s", schemaIdent, tableIdent)
	tableFQNText := fmt.Sprintf("%s.%s", policy.Schema, policy.Table)
	command := strings.ToUpper(strings.TrimSpace(string(policy.Command)))
	if command == "" {
		command = string(objects.PolicyCommandAll)
	}
	alter := fmt.Sprintf("ALTER POLICY %q ON %s.%s", updatePolicyParams.Name, schemaIdent, tableIdent)
	grantAccessTables := []string{}

	var nameSql, definitionSql, checkSql, rolesSql string
	for _, ut := range updatePolicyParams.ChangeItems {
		switch ut {
		case objects.UpdatePolicyName:
			if policy.Name != "" {
				nameSql = fmt.Sprintf("%s RENAME TO %s;", alter, policy.Name)
			}
		case objects.UpdatePolicyCheck:
			if policy.Check == nil || (policy.Check != nil && *policy.Check == "") {
				defaultCheck := "true"
				policy.Check = &defaultCheck
			}
			checkSql = fmt.Sprintf("%s WITH CHECK (%s);", alter, *policy.Check)
		case objects.UpdatePolicyDefinition:
			if command != string(objects.PolicyCommandInsert) && policy.Definition != "" {
				definitionSql = fmt.Sprintf("%s USING (%s);", alter, policy.Definition)
			}
		case objects.UpdatePolicyRoles:
			if len(policy.Roles) > 0 {
				quotedRoles := make([]string, 0, len(policy.Roles))
				for _, role := range policy.Roles {
					quotedRoles = append(quotedRoles, pq.QuoteIdentifier(role))
				}
				rolesSql = fmt.Sprintf("%s TO %s;", alter, strings.Join(quotedRoles, ", "))
			} else {
				rolesSql = fmt.Sprintf("%s TO PUBLIC;", alter)
			}

			for _, role := range policy.Roles {
				for _, privilege := range privilegesForCommand(command) {
					grantAccessTables = append(grantAccessTables, fmt.Sprintf(`
				IF NOT HAS_TABLE_PRIVILEGE('%s', '%s', '%s') THEN
					GRANT %s ON %s TO %s;
				END IF;
			`, role, tableFQNText, privilege, privilege, tableFQN, pq.QuoteIdentifier(role)))
				}
			}
		}
	}

	return fmt.Sprintf("DO $$ BEGIN %s %s %s %s %s END $$;", definitionSql, checkSql, rolesSql, nameSql, strings.Join(grantAccessTables, "\n"))
}

func BuildDeletePolicyQuery(policy objects.Policy) string {
	revokeAccessTables := []string{}
	privileges := privilegesForCommand(strings.ToUpper(strings.TrimSpace(string(policy.Command))))
	for _, role := range policy.Roles {
		for _, privilege := range privileges {
			revokeAccessTables = append(revokeAccessTables, fmt.Sprintf(`
			IF HAS_TABLE_PRIVILEGE('%s', '%s.%s', '%s') THEN
				REVOKE %s ON %s.%s FROM %s;
			END IF;
		`, role, policy.Schema, policy.Table, privilege, privilege, policy.Schema, policy.Table, role))
		}
	}

	revokeAccessQuery := fmt.Sprintf(`
	DO $$
	BEGIN
		DROP POLICY %q ON %s.%s;
	%s
	
	END $$;
	`, policy.Name, policy.Schema, policy.Table, strings.Join(revokeAccessTables, "\n"))

	return revokeAccessQuery
}

func privilegesForCommand(command string) []string {
	switch command {
	case string(objects.PolicyCommandAll):
		return []string{
			string(objects.PolicyCommandSelect),
			string(objects.PolicyCommandInsert),
			string(objects.PolicyCommandUpdate),
			string(objects.PolicyCommandDelete),
		}
	default:
		return []string{command}
	}
}
