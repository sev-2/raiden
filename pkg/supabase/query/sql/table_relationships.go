package sql

import "fmt"

var GetTableRelationshipsQuery = `
-- Adapted from
-- https://github.com/PostgREST/postgrest/blob/f9f0f79fa914ac00c11fbf7f4c558e14821e67e2/src/PostgREST/SchemaCache.hs#L722
WITH
pks_uniques_cols AS (
  SELECT
    connamespace,
    conrelid,
    jsonb_agg(column_info.cols) as cols
  FROM pg_constraint
  JOIN lateral (
    SELECT array_agg(cols.attname order by cols.attnum) as cols
    FROM ( select unnest(conkey) as col) _
    JOIN pg_attribute cols on cols.attrelid = conrelid and cols.attnum = col
  ) column_info ON TRUE
  WHERE
    contype IN ('p', 'u') and
    connamespace::regnamespace::text <> 'pg_catalog'
  GROUP BY connamespace, conrelid
)
SELECT
  traint.conname AS foreign_key_name,
  ns1.nspname AS schema,
  tab.relname AS relation,
  column_info.cols AS columns,
  ns2.nspname AS referenced_schema,
  other.relname AS referenced_relation,
  column_info.refs AS referenced_columns,
  (column_info.cols IN (SELECT * FROM jsonb_array_elements(pks_uqs.cols))) AS is_one_to_one
FROM pg_constraint traint
JOIN LATERAL (
  SELECT
    jsonb_agg(cols.attname order by ord) AS cols,
    jsonb_agg(refs.attname order by ord) AS refs
  FROM unnest(traint.conkey, traint.confkey) WITH ORDINALITY AS _(col, ref, ord)
  JOIN pg_attribute cols ON cols.attrelid = traint.conrelid AND cols.attnum = col
  JOIN pg_attribute refs ON refs.attrelid = traint.confrelid AND refs.attnum = ref
) AS column_info ON TRUE
JOIN pg_namespace ns1 ON ns1.oid = traint.connamespace
JOIN pg_class tab ON tab.oid = traint.conrelid
JOIN pg_class other ON other.oid = traint.confrelid
JOIN pg_namespace ns2 ON ns2.oid = other.relnamespace
LEFT JOIN pks_uniques_cols pks_uqs ON pks_uqs.connamespace = traint.connamespace AND pks_uqs.conrelid = traint.conrelid
WHERE traint.contype = 'f'
AND traint.conparentid = 0
`

var GetTableRelationshipActionsQuery = `
SELECT 
  con.oid as id, 
  con.conname as constraint_name, 
  con.confdeltype as deletion_action,
  con.confupdtype as update_action,
  rel.oid as source_id,
  nsp.nspname as source_schema, 
  rel.relname as source_table, 
  (
    SELECT 
      array_agg(
        att.attname 
        ORDER BY 
          un.ord
      ) 
    FROM 
      unnest(con.conkey) WITH ORDINALITY un (attnum, ord) 
      INNER JOIN pg_attribute att ON att.attnum = un.attnum 
    WHERE 
      att.attrelid = rel.oid
  ) source_columns, 
  frel.oid as target_id,
  fnsp.nspname as target_schema, 
  frel.relname as target_table, 
  (
    SELECT 
      array_agg(
        att.attname 
        ORDER BY 
          un.ord
      ) 
    FROM 
      unnest(con.confkey) WITH ORDINALITY un (attnum, ord) 
      INNER JOIN pg_attribute att ON att.attnum = un.attnum 
    WHERE 
      att.attrelid = frel.oid
  ) target_columns 
FROM 
  pg_constraint con 
  INNER JOIN pg_class rel ON rel.oid = con.conrelid 
  INNER JOIN pg_namespace nsp ON nsp.oid = rel.relnamespace 
  INNER JOIN pg_class frel ON frel.oid = con.confrelid 
  INNER JOIN pg_namespace fnsp ON fnsp.oid = frel.relnamespace 
WHERE 
  con.contype = 'f'
`

func GenerateGetTableRelationshipActionsQuery(schema string) string {
	if len(schema) == 0 {
		schema = "public"
	}

	filteredSql := GetTableRelationshipActionsQuery + " AND nsp.nspname = %s"
	schemaFilter := fmt.Sprintf("'%s'", schema)

	return fmt.Sprintf(filteredSql, schemaFilter)
}
