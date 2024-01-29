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
