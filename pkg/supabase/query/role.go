package query

import (
	"fmt"
	"strings"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/supabase/objects"
)

func BuildCreateRoleQuery(role objects.Role) string {
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
				configStrings = append(configStrings, fmt.Sprintf("ALTER ROLE %s SET %s = %s;", role.Name, k, v))
			}
		}
		configClause = strings.Join(configStrings, " ")
	}

	return fmt.Sprintf(`
	BEGIN;
	do $$
	BEGIN
		IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = '%s') THEN
			CREATE ROLE %s WITH %s;
		END IF;
	END $$;
	%s
	COMMIT;`,
		role.Name, role.Name, strings.Join(createRolClauses, "\n"),
		configClause,
	)
}

func BuildUpdateRoleQuery(newRole objects.Role, updateRoleParam objects.UpdateRoleParam) string {
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

	return fmt.Sprintf(`
		BEGIN; %s %s %s COMMIT;
	`, updateRoleClause, configClause, nameClause)
}

func BuildDeleteRoleQuery(role objects.Role) string {
	return fmt.Sprintf("DROP ROLE %s;", role.Name)
}
