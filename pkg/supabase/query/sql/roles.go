package sql

import (
	"fmt"
	"strings"
)

var GetRolesQuery = `
-- TODO: Consider using pg_authid vs. pg_roles for unencrypted password field
SELECT
  oid :: int8 AS id,
  rolname AS name,
  rolsuper AS is_superuser,
  rolcreatedb AS can_create_db,
  rolcreaterole AS can_create_role,
  rolinherit AS inherit_role,
  rolcanlogin AS can_login,
  rolreplication AS is_replication_role,
  rolbypassrls AS can_bypass_rls,
  (
    SELECT
      COUNT(*)
    FROM
      pg_stat_activity
    WHERE
      pg_roles.rolname = pg_stat_activity.usename
  ) AS active_connections,
  CASE WHEN rolconnlimit = -1 THEN current_setting('max_connections') :: int8
       ELSE rolconnlimit
  END AS connection_limit,
  rolpassword AS password,
  rolvaliduntil AS valid_until,
  rolconfig AS config
FROM
  pg_roles
`

var getRoleMembershipsQuery = `
WITH schema_roles AS (%s),
role_inheritance AS (
  SELECT
    role.oid::int8    AS parent_id,
    role.rolname      AS parent_role,
    inherit.oid::int8  AS inherit_id,
    inherit.rolname    AS inherit_role_name
  FROM pg_auth_members m
  JOIN pg_roles inherit ON m.member = inherit.oid
  JOIN pg_roles role   ON m.roleid = role.oid
)
SELECT DISTINCT
  i.parent_id,
  i.parent_role,
  i.inherit_id,
  i.inherit_role_name,
  sr.schema_name
FROM role_inheritance i
LEFT JOIN schema_roles sr
  ON sr.role_name IN (i.parent_role, i.inherit_role_name)
WHERE sr.schema_name IS NOT NULL
ORDER BY i.parent_role, i.inherit_role_name;
`

func GenerateGetRoleMembershipsQuery(includedSchema []string) string {
	var filterSchema []string

	var schema_roles = `
SELECT
  r.rolname AS role_name,
  n.nspname AS schema_name
FROM pg_namespace n
JOIN LATERAL unnest(n.nspacl) acl(entry) ON TRUE
JOIN pg_roles r ON acl::text LIKE '%' || r.rolname || '%'
  `

	if len(includedSchema) > 0 {
		for _, s := range includedSchema {
			filterSchema = append(filterSchema, fmt.Sprintf("'%s'", s))
		}

		schema_roles += fmt.Sprintf("WHERE n.nspname = ANY(ARRAY[%s]) ", strings.Join(filterSchema, ","))
	}

	return fmt.Sprintf(getRoleMembershipsQuery, schema_roles)
}
