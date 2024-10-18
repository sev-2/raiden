package sql

import "fmt"

var GetIndexesQuery = `
SELECT 
    schemaname AS "schema",
    tablename AS "table",
    indexname AS "name",
    indexdef AS "definition"
FROM 
    pg_indexes
`

func GenerateGetIndexQuery(schema string) string {
	if len(schema) == 0 {
		schema = "public"
	}

	filteredSql := GetIndexesQuery + " WHERE schemaname = %s"
	schemaFilter := fmt.Sprintf("'%s'", schema)

	return fmt.Sprintf(filteredSql, schemaFilter)
}
