package query

import (
	"errors"
	"fmt"
	"strings"

	"github.com/lib/pq"
	"github.com/sev-2/raiden/pkg/supabase/objects"
)

type FunctionAction string

const (
	FunctionActionCreate FunctionAction = "create"
	FunctionActionUpdate FunctionAction = "update"
	FunctionActionDelete FunctionAction = "delete"
)

func BuildFunctionQuery(action FunctionAction, fn *objects.Function) (string, error) {
	if fn == nil {
		return "", errors.New("function payload is required")
	}

	schemaIdent := pq.QuoteIdentifier(fn.Schema)
	nameIdent := pq.QuoteIdentifier(fn.Name)
	dropStmt := buildDropFunctionStatement(schemaIdent, nameIdent, fn)
	createStmt := strings.TrimSpace(fn.CompleteStatement)
	if createStmt == "" && action != FunctionActionDelete {
		return "", errors.New("function complete statement is required")
	}
	if createStmt != "" && !strings.HasSuffix(createStmt, ";") {
		createStmt += ";"
	}

	switch action {
	case FunctionActionCreate:
		return createStmt, nil
	case FunctionActionDelete:
		return dropStmt, nil
	case FunctionActionUpdate:
		return fmt.Sprintf("BEGIN; %s %s COMMIT;", dropStmt, createStmt), nil

	default:
		return "", fmt.Errorf("generate function sql with type '%s' is not available", action)
	}
}

func buildDropFunctionStatement(schemaIdent, nameIdent string, fn *objects.Function) string {
	signature := strings.TrimSpace(fn.IdentityArgumentTypes)
	if signature == "" {
		signature = strings.TrimSpace(fn.ArgumentTypes)
	}
	if signature == "" {
		signature = ""
	}

	if signature == "" {
		return fmt.Sprintf("DROP FUNCTION IF EXISTS %s.%s();", schemaIdent, nameIdent)
	}

	return fmt.Sprintf("DROP FUNCTION IF EXISTS %s.%s(%s);", schemaIdent, nameIdent, signature)
}
