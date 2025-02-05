package query

import (
	"fmt"
	"strings"

	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/supabase/query/sql"
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
	enums := []string{}
	for _, e := range dataType.Enums {
		enums = append(enums, sql.Literal(e))
	}
	return fmt.Sprintf(`CREATE TYPE %s.%s AS ENUM (%s);`, dataType.Schema, dataType.Name, strings.Join(enums, ","))
}

func BuildDeleteTypeQuery(dataType *objects.Type) string {
	if dataType == nil {
		return ""
	}
	return fmt.Sprintf(`DROP TYPE IF EXISTS %s.%s CASCADE;`, dataType.Schema, dataType.Name)
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
