package sql

import (
	"fmt"
	"strings"
)

func coalesceRowsToArray(table, condition string) string {
	return fmt.Sprintf("COALESCE((SELECT JSON_AGG(row_to_json(%s)) FROM %s WHERE %s), '[]') as %s", table, table, condition, table)
}

func filterByList(include, exclude, defaultExclude []string) string {
	if defaultExclude != nil {
		exclude = append(defaultExclude, exclude...)
	}
	if len(include) > 0 {
		return fmt.Sprintf("IN (%s)", strings.Join(mapToLiterals(include), ","))
	} else if len(exclude) > 0 {
		return fmt.Sprintf("NOT IN (%s)", strings.Join(mapToLiterals(exclude), ","))
	}
	return ""
}

func literal(s string) string {
	return fmt.Sprintf("'%s'", s)
}

func mapToLiterals(strings []string) []string {
	result := make([]string, len(strings))
	for i, s := range strings {
		result[i] = literal(s)
	}
	return result
}
