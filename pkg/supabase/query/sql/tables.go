package sql

import (
	"strings"
	"text/template"
)

var GetTablesQuery = `
SELECT
  c.oid :: int8 AS id,
  nc.nspname AS schema,
  c.relname AS name,
  c.relrowsecurity AS rls_enabled,
  c.relforcerowsecurity AS rls_forced,
  CASE
    WHEN c.relreplident = 'd' THEN 'DEFAULT'
    WHEN c.relreplident = 'i' THEN 'INDEX'
    WHEN c.relreplident = 'f' THEN 'FULL'
    ELSE 'NOTHING'
  END AS replica_identity,
  pg_total_relation_size(format('%I.%I', nc.nspname, c.relname)) :: int8 AS bytes,
  pg_size_pretty(
    pg_total_relation_size(format('%I.%I', nc.nspname, c.relname))
  ) AS size,
  pg_stat_get_live_tuples(c.oid) AS live_rows_estimate,
  pg_stat_get_dead_tuples(c.oid) AS dead_rows_estimate,
  obj_description(c.oid) AS comment,
  coalesce(pk.primary_keys, '[]') as primary_keys,
  coalesce(
    jsonb_agg(relationships) filter (where relationships is not null),
    '[]'
  ) as relationships
FROM
  pg_namespace nc
  JOIN pg_class c ON nc.oid = c.relnamespace
  left join (
    select
      table_id,
      jsonb_agg(_pk.*) as primary_keys
    from (
      select
        n.nspname as schema,
        c.relname as table_name,
        a.attname as name,
        c.oid :: int8 as table_id
      from
        pg_index i,
        pg_class c,
        pg_attribute a,
        pg_namespace n
      where
        i.indrelid = c.oid
        and c.relnamespace = n.oid
        and a.attrelid = c.oid
        and a.attnum = any (i.indkey)
        and i.indisprimary
    ) as _pk
    group by table_id
  ) as pk
  on pk.table_id = c.oid
  left join (
    select
      c.oid :: int8 as id,
      c.conname as constraint_name,
      nsa.nspname as source_schema,
      csa.relname as source_table_name,
      sa.attname as source_column_name,
      nta.nspname as target_table_schema,
      cta.relname as target_table_name,
      ta.attname as target_column_name
    from
      pg_constraint c
    join (
      pg_attribute sa
      join pg_class csa on sa.attrelid = csa.oid
      join pg_namespace nsa on csa.relnamespace = nsa.oid
    ) on sa.attrelid = c.conrelid and sa.attnum = any (c.conkey)
    join (
      pg_attribute ta
      join pg_class cta on ta.attrelid = cta.oid
      join pg_namespace nta on cta.relnamespace = nta.oid
    ) on ta.attrelid = c.confrelid and ta.attnum = any (c.confkey)
    where
      c.contype = 'f'
  ) as relationships
  on (relationships.source_schema = nc.nspname and relationships.source_table_name = c.relname)
  or (relationships.target_table_schema = nc.nspname and relationships.target_table_name = c.relname)
WHERE
  c.relkind IN ('r', 'p')
  AND NOT pg_is_other_temp_schema(nc.oid)
  AND (
    pg_has_role(c.relowner, 'USAGE')
    OR has_table_privilege(
      c.oid,
      'SELECT, INSERT, UPDATE, DELETE, TRUNCATE, REFERENCES, TRIGGER'
    )
    OR has_any_column_privilege(c.oid, 'SELECT, INSERT, UPDATE, REFERENCES')
  )
group by
  c.oid,
  c.relname,
  c.relrowsecurity,
  c.relforcerowsecurity,
  c.relreplident,
  nc.nspname,
  pk.primary_keys
`

const tablesQueryTemplate = `
WITH tables AS ({{.TablesSQL}})
{{if .IncludeColumns}}
  , columns AS ({{.ColumnsSQL}})
{{end}}
SELECT
  *
{{if .IncludeColumns}}
  , {{coalesceRowsToArray "columns" "columns.table_id = tables.id"}}
{{end}}
FROM tables 
{{if .IncludeSchemas }}
where schema {{.FilterSQL}}
{{end}}
`

const tableQueryTemplate = `
WITH tables AS ({{.TablesSQL}})
{{if .IncludeColumns}}
  , columns AS ({{.ColumnsSQL}})
{{end}}
SELECT
  *
{{if .IncludeColumns}}
  , {{coalesceRowsToArray "columns" "columns.table_id = tables.id"}}
{{end}}
FROM tables 
where schema {{ .FilterSchemaSQL }} AND name {{ .FilterNameSQL }} LIMIT 1
`

func GenerateGetTablesQuery(includeSchemas []string, includeColumn bool) (string, error) {
	tmpl, err := template.New("enrichedTablesSQL").
		Funcs(template.FuncMap{
			"coalesceRowsToArray": coalesceRowsToArray,
		}).
		Parse(tablesQueryTemplate)

	if err != nil {
		return "", err
	}

	var result strings.Builder
	err = tmpl.Execute(&result, map[string]interface{}{
		"TablesSQL":      GetTablesQuery,
		"ColumnsSQL":     GetColumnsQuery,
		"IncludeColumns": includeColumn,
		"IncludeSchemas": len(includeSchemas) > 0,
		"FilterSQL":      filterByList(includeSchemas, nil, nil),
	})

	if err != nil {
		return "", err
	}

	return result.String(), nil
}

func GenerateGetTableQuery(tableName string, schema string, includeColumn bool) (string, error) {
	tmpl, err := template.New("enrichedTablesSQL").
		Funcs(template.FuncMap{
			"coalesceRowsToArray": coalesceRowsToArray,
		}).
		Parse(tableQueryTemplate)

	if err != nil {
		return "", err
	}

	var result strings.Builder
	err = tmpl.Execute(&result, map[string]interface{}{
		"TablesSQL":       GetTablesQuery,
		"ColumnsSQL":      GetColumnsQuery,
		"IncludeColumns":  includeColumn,
		"FilterSchemaSQL": filterByList([]string{schema}, nil, nil),
		"FilterNameSQL":   filterByList([]string{tableName}, nil, nil),
	})

	if err != nil {
		return "", err
	}

	return result.String(), nil
}
