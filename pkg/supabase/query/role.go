package query

import (
	"fmt"
	"strings"

	"github.com/lib/pq"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/supabase/objects"
)

func BuildCreateRoleQuery(role objects.Role) string {
	roleIdent := pq.QuoteIdentifier(role.Name)
	roleLiteral := pq.QuoteLiteral(role.Name)

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

	// isSuperuserClause := "NOSUPERUSER"
	// if role.IsSuperuser {
	// 	isSuperuserClause = "SUPERUSER"
	// }
	// createRolClauses = append(createRolClauses, isSuperuserClause)

	// isReplicationRoleClause := "NOREPLICATION"
	// if role.IsReplicationRole {
	// 	isReplicationRoleClause = "REPLICATION"
	// }
	// createRolClauses = append(createRolClauses, isReplicationRoleClause)

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
				configStrings = append(configStrings, fmt.Sprintf("ALTER ROLE %s SET %s = %s;", roleIdent, pq.QuoteIdentifier(k), pq.QuoteLiteral(fmt.Sprint(v))))
			}
		}
		configClause = strings.Join(configStrings, " ")
	}

	if configClause == "" {
		configClause = ""
	}

	return fmt.Sprintf(`
	BEGIN;
	DO $$
	BEGIN
		IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = %s) THEN
			CREATE ROLE %s WITH %s;
			GRANT %s TO authenticator;
			GRANT anon TO %s;
		END IF;
	END $$;
	%s
	GRANT %s TO authenticator;
	COMMIT;`,
		roleLiteral,
		roleIdent,
		strings.Join(createRolClauses, "\n"),
		roleIdent,
		roleIdent,
		configClause,
		roleIdent,
	)
}

func BuildUpdateRoleQuery(newRole objects.Role, updateRoleParam objects.UpdateRoleParam) string {
	alter := fmt.Sprintf("ALTER ROLE %s", pq.QuoteIdentifier(updateRoleParam.OldData.Name))
	newRoleIdent := pq.QuoteIdentifier(newRole.Name)

	var updateRoleClause, nameClause, configClause string
	var updateRoleClauses []string

	for _, item := range updateRoleParam.ChangeItems {
		switch item {
		case objects.UpdateConnectionLimit:
			updateRoleClauses = append(updateRoleClauses, fmt.Sprintf("CONNECTION LIMIT %d", newRole.ConnectionLimit))
		case objects.UpdateRoleName:
			if updateRoleParam.OldData.Name != newRole.Name && newRole.Name != "" {
				nameClause = fmt.Sprintf("ALTER ROLE %s RENAME TO %s;", pq.QuoteIdentifier(updateRoleParam.OldData.Name), newRoleIdent)
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
			if newRole.ValidUntil != nil {
				validUntilClause := fmt.Sprintf("VALID UNTIL '%s'", newRole.ValidUntil.Format(raiden.DefaultRoleValidUntilLayout))
				updateRoleClauses = append(updateRoleClauses, validUntilClause)
			}
		case objects.UpdateRoleConfig:
			var configStrings []string
			for k, v := range newRole.Config {
				if k != "" && v != "" {
					configStrings = append(configStrings, fmt.Sprintf("%s SET %s = %s;", alter, pq.QuoteIdentifier(k), pq.QuoteLiteral(fmt.Sprint(v))))
				}
			}
			configClause = strings.Join(configStrings, "\n")
		}
	}

	if len(updateRoleClauses) > 0 {
		updateRoleClause = fmt.Sprintf("%s %s;", alter, strings.Join(updateRoleClauses, "\n"))
	}

	return fmt.Sprintf("BEGIN; %s %s %s COMMIT;", updateRoleClause, configClause, nameClause)
}

func BuildDeleteRoleQuery(role objects.Role) string {
	roleLiteral := pq.QuoteLiteral(role.Name)
	return fmt.Sprintf(`
DO $$
DECLARE
	rec RECORD;
	role_exists BOOLEAN;
	role_name TEXT := %[1]s;
	target_owner TEXT := 'postgres';
BEGIN
	SELECT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = role_name)
	INTO role_exists;

	IF NOT role_exists THEN
		RAISE EXCEPTION 'Role "%%" does not exist.', role_name;
	END IF;

	RAISE NOTICE 'Starting cleanup for role: %%', role_name;

	-- Revoke privileges only from safe, non-system schemas
	FOR rec IN
		SELECT nspname AS schema_name
		FROM pg_namespace
		WHERE (has_schema_privilege(role_name, nspname, 'USAGE')
		       OR has_schema_privilege(role_name, nspname, 'CREATE'))
		  AND nspname NOT IN (
		      'pg_catalog', 'information_schema', 'net',
		      'graphql_public', 'storage', 'auth', 'extensions',
		      'supabase_functions', 'realtime', 'pg_toast'
		  )
	LOOP
		BEGIN
			EXECUTE format('REVOKE ALL PRIVILEGES ON SCHEMA %%I FROM %%I;', rec.schema_name, role_name);
			EXECUTE format('REVOKE ALL PRIVILEGES ON ALL TABLES IN SCHEMA %%I FROM %%I;', rec.schema_name, role_name);
			EXECUTE format('REVOKE ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA %%I FROM %%I;', rec.schema_name, role_name);
			EXECUTE format('REVOKE ALL PRIVILEGES ON ALL FUNCTIONS IN SCHEMA %%I FROM %%I;', rec.schema_name, role_name);
		EXCEPTION WHEN OTHERS THEN
			RAISE NOTICE 'Skipped schema %%: %%', rec.schema_name, SQLERRM;
		END;
	END LOOP;

	-- Try reassign & drop owned (skip errors if blocked)
	BEGIN
		EXECUTE format('REASSIGN OWNED BY %%I TO %%I;', role_name, target_owner);
	EXCEPTION WHEN OTHERS THEN
		RAISE NOTICE 'Cannot reassign owned: %%', SQLERRM;
	END;

	BEGIN
		EXECUTE format('DROP OWNED BY %%I CASCADE;', role_name);
	EXCEPTION WHEN OTHERS THEN
		RAISE NOTICE 'Cannot drop owned: %%', SQLERRM;
	END;

	-- Try drop role itself
	BEGIN
		EXECUTE format('REVOKE %%I FROM authenticator;', role_name);
		EXECUTE format('REVOKE anon FROM %%I;', role_name);
		EXECUTE format('DROP ROLE %%I;', role_name);
	EXCEPTION WHEN OTHERS THEN
		RAISE NOTICE 'Cannot drop role: %%', SQLERRM;
	END;
END $$;`, roleLiteral)
}

func BuildRoleInheritQuery(roleName string, inheritRoleName string, action objects.UpdateRoleInheritType) (string, error) {
	if roleName == "" {
		return "", fmt.Errorf("role name is required")
	}

	if inheritRoleName == "" {
		return "", fmt.Errorf("inherit role name is required")
	}

	quotedRole := pq.QuoteIdentifier(roleName)
	quotedInherit := pq.QuoteIdentifier(inheritRoleName)

	switch action {
	case objects.UpdateRoleInheritGrant:
		return fmt.Sprintf("GRANT %s TO %s;", quotedInherit, quotedRole), nil
	case objects.UpdateRoleInheritRevoke:
		return fmt.Sprintf("REVOKE %s FROM %s;", quotedInherit, quotedRole), nil
	default:
		return "", fmt.Errorf("unsupported role inherit action %s", action)
	}
}
