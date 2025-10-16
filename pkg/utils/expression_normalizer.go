package utils

import (
	"regexp"
	"strings"
)

// ExpressionQualifier describes schema/table prefixes that Postgres may
// inject into expressions.
type ExpressionQualifier struct {
	Schema string
	Table  string
}

var (
	identifierParenPattern = regexp.MustCompile(`\(("?[a-zA-Z_][a-zA-Z0-9_\.]*"?)\)`)
	typeCastPattern        = regexp.MustCompile(`::("?[a-zA-Z_][a-zA-Z0-9_\.]*"?)`)
)

// NormalizeExpression canonicalizes a SQL boolean expression so formatting
// differences introduced by Postgres (casts, redundant parentheses, quoted
// identifiers) do not trigger spurious diffs.
func NormalizeExpression(expr string, qualifiers ...ExpressionQualifier) string {
	return normalizeExpression(expr, qualifiers...)
}

// NormalizeOptionalExpression handles optional expressions, returning the
// normalized value and whether a clause was present.
func NormalizeOptionalExpression(expr *string, qualifiers ...ExpressionQualifier) (string, bool) {
	if expr == nil {
		return "", false
	}
	return normalizeExpression(*expr, qualifiers...), true
}

func normalizeExpression(expr string, qualifiers ...ExpressionQualifier) string {
	trimmed := strings.TrimSpace(expr)
	for len(trimmed) >= 2 && trimmed[0] == '(' && trimmed[len(trimmed)-1] == ')' && enclosedByOuterParentheses(trimmed) {
		trimmed = strings.TrimSpace(trimmed[1 : len(trimmed)-1])
	}
	trimmed = strings.ReplaceAll(trimmed, "\"", "")
	trimmed = typeCastPattern.ReplaceAllString(trimmed, "")
	trimmed = collapseWhitespace(trimmed)
	for {
		next := identifierParenPattern.ReplaceAllString(trimmed, "$1")
		if next == trimmed {
			break
		}
		trimmed = next
	}
	trimmed = removeExpressionQualifiers(trimmed, qualifiers...)
	trimmed = collapseWhitespace(strings.TrimSpace(trimmed))
	return canonicalizeSimpleEquality(trimmed)
}

func removeExpressionQualifiers(expr string, qualifiers ...ExpressionQualifier) string {
	result := expr
	for _, q := range qualifiers {
		table := strings.TrimSpace(strings.ToLower(q.Table))
		schema := strings.TrimSpace(strings.ToLower(q.Schema))
		if table == "" {
			continue
		}
		patterns := []string{table + "."}
		if schema != "" {
			patterns = append(patterns, schema+"."+table+".")
			patterns = append(patterns, schema+".")
		}
		for _, p := range patterns {
			re := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(p))
			result = re.ReplaceAllString(result, "")
		}
	}
	return result
}

func canonicalizeSimpleEquality(expr string) string {
	eqIndex := strings.Index(expr, "=")
	if eqIndex <= 0 {
		return expr
	}
	if strings.Contains(expr, " AND ") || strings.Contains(expr, " OR ") || strings.Contains(expr, " NOT ") {
		return expr
	}
	if strings.ContainsAny(expr, "<>!") {
		return expr
	}
	lhs := strings.TrimSpace(expr[:eqIndex])
	rhs := strings.TrimSpace(expr[eqIndex+1:])
	if lhs == "" || rhs == "" {
		return expr
	}
	if strings.Compare(lhs, rhs) > 0 {
		lhs, rhs = rhs, lhs
	}
	return lhs + " = " + rhs
}

func enclosedByOuterParentheses(expr string) bool {
	depth := 0
	for i, ch := range expr {
		switch ch {
		case '(':
			depth++
		case ')':
			depth--
			if depth < 0 {
				return false
			}
			if depth == 0 && i < len(expr)-1 {
				return false
			}
		}
	}
	return depth == 0
}

func collapseWhitespace(expr string) string {
	if expr == "" {
		return expr
	}
	return strings.Join(strings.Fields(expr), " ")
}
