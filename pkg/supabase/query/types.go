package query

import (
	"fmt"
	"strings"

	"github.com/lib/pq"
	"github.com/sev-2/raiden/pkg/supabase/objects"
)

type TypeAction string

const (
	TypeActionCreate TypeAction = "create"
	TypeActionUpdate TypeAction = "update"
	TypeActionDelete TypeAction = "delete"
)

func BuildCreateTypeQuery(dataType *objects.Type) string {
	if dataType == nil {
		return ""
	}
	schemaIdent := pq.QuoteIdentifier(dataType.Schema)
	nameIdent := pq.QuoteIdentifier(dataType.Name)
	enums := []string{}
	for _, e := range dataType.Enums {
		enums = append(enums, pq.QuoteLiteral(e))
	}
	return fmt.Sprintf(`CREATE TYPE IF NOT EXISTS %s.%s AS ENUM (%s);`, schemaIdent, nameIdent, strings.Join(enums, ","))
}

func BuildDeleteTypeQuery(dataType *objects.Type) string {
	if dataType == nil {
		return ""
	}
	schemaIdent := pq.QuoteIdentifier(dataType.Schema)
	nameIdent := pq.QuoteIdentifier(dataType.Name)
	return fmt.Sprintf(`DROP TYPE IF EXISTS %s.%s CASCADE;`, schemaIdent, nameIdent)
}

func BuildTypeQuery(action TypeAction, dataType *objects.Type) (string, error) {
	switch action {
	case TypeActionCreate:
		return BuildCreateTypeQuery(dataType), nil
	case TypeActionDelete:
		return BuildDeleteTypeQuery(dataType), nil
	case TypeActionUpdate:
		return fmt.Sprintf(`
			BEGIN; 
				%s 
				%s  
			COMMIT;
		`, BuildDeleteTypeQuery(dataType), BuildCreateTypeQuery(dataType)), nil

	default:
		return "", fmt.Errorf("generate type sql with action '%s' is not available", action)
	}
}
