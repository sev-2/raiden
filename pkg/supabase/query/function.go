package query

import (
	"fmt"

	"github.com/sev-2/raiden/pkg/supabase/objects"
)

type FunctionAction string

const (
	FunctionActionCreate FunctionAction = "create"
	FunctionActionUpdate FunctionAction = "update"
	FunctionActionDelete FunctionAction = "delete"
)

func BuildFunctionQuery(action FunctionAction, fn *objects.Function) (string, error) {
	switch action {
	case FunctionActionCreate:
		return fn.CompleteStatement + ";", nil
	case FunctionActionDelete:
		return fmt.Sprintf("DROP FUNCTION %s.%s;", fn.Schema, fn.Name), nil
	case FunctionActionUpdate:
		return fmt.Sprintf(`
			BEGIN; 
				%s 
				%s  
			COMMIT;
		`, fmt.Sprintf("DROP FUNCTION %s.%s;", fn.Schema, fn.Name), fn.CompleteStatement+";"), nil

	default:
		return "", fmt.Errorf("generate function sql with type '%s' is not available", action)
	}
}
