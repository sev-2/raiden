package postgres

import (
	"fmt"
	"strings"
)

// Convert postgres type to golang type
func ToGoType(pgType string, isNullable bool) (goType string) {
	switch strings.ToLower(pgType) {
	case "bigint":
		goType = "int64"
	case "integer", "serial":
		goType = "int32"
	case "smallint":
		goType = "int16"
	case "real":
		goType = "float32"
	case "double precision":
		goType = "float64"
	case "numeric":
		goType = "float64"
	case "boolean":
		goType = "bool"
	case "text", "character varying":
		goType = "string"
	case "date", "timestamp with time zone", "timestamp without time zone":
		goType = "time.Time"
	case "uuid":
		goType = "uuid.UUID"
	case "json", "jsonb":
		goType = "json.RawMessage"
	default:
		goType = "any"
	}

	if isNullable {
		goType = fmt.Sprintf("*%s", goType)
	}

	return
}

func ToPostgresType(goType string) (pgType string) {
	// Map Go types to PostgreSQL types
	switch goType {
	case "int64":
		pgType = "bigint"
	case "int32", "int":
		pgType = "integer"
	case "int16":
		pgType = "smallint"
	case "float32":
		pgType = "real"
	case "float64":
		pgType = "double precision"
	case "bool":
		pgType = "boolean"
	case "string":
		pgType = "text"
	case "time.Time":
		pgType = "timestamp with time zone"
	case "uuid.UUID":
		pgType = "uuid"
	case "json.RawMessage":
		pgType = "jsonb"
	default:
		pgType = "varchar"
	}

	return
}
