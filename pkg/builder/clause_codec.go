package builder

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

type ClauseQualifier struct {
	Schema string
	Table  string
}

var (
	identifierParenPattern = regexp.MustCompile(`\(("?[a-zA-Z_][a-zA-Z0-9_\.]*"?)\)`)
	typeCastPattern        = regexp.MustCompile(`::("?[a-zA-Z_][a-zA-Z0-9_\.]*"?)`)
)

func NormalizeClauseSQL(sql string, qualifiers ...ClauseQualifier) string {
	return normalizeClause(sql, qualifiers...)
}

func MarshalClause(c Clause) string {
	return strings.TrimSpace(string(c))
}

func UnmarshalClause(sql string, qualifiers ...ClauseQualifier) (Clause, string, bool) {
	normalized := NormalizeClauseSQL(sql, qualifiers...)
	if strings.TrimSpace(normalized) == "" {
		return Clause(""), "b.Clause(\"\")", true
	}
	clause, code, ok := parseClause(normalized)
	return clause, code, ok
}

func normalizeClause(expr string, qualifiers ...ClauseQualifier) string {
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
	trimmed = removeQualifiers(trimmed, qualifiers...)
	trimmed = collapseWhitespace(strings.TrimSpace(trimmed))
	return canonicalizeSimpleEquality(trimmed)
}

func parseClause(expr string) (Clause, string, bool) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return Clause(""), "b.Clause(\"\")", true
	}
	if enclosedByOuterParentheses(expr) {
		return parseClause(strings.TrimSpace(expr[1 : len(expr)-1]))
	}
	if parts := splitByLogical(expr, "OR"); parts != nil {
		clauses := make([]Clause, 0, len(parts))
		codes := make([]string, 0, len(parts))
		for _, part := range parts {
			cl, code, ok := parseClause(part)
			if !ok {
				return "", "", false
			}
			clauses = append(clauses, cl)
			codes = append(codes, code)
		}
		if len(clauses) == 1 {
			return clauses[0], codes[0], true
		}
		return Or(clauses...), "b.Or(" + strings.Join(codes, ", ") + ")", true
	}
	if parts := splitByLogical(expr, "AND"); parts != nil {
		clauses := make([]Clause, 0, len(parts))
		codes := make([]string, 0, len(parts))
		for _, part := range parts {
			cl, code, ok := parseClause(part)
			if !ok {
				return "", "", false
			}
			clauses = append(clauses, cl)
			codes = append(codes, code)
		}
		if len(clauses) == 1 {
			return clauses[0], codes[0], true
		}
		return And(clauses...), "b.And(" + strings.Join(codes, ", ") + ")", true
	}
	upper := strings.ToUpper(expr)
	if strings.HasPrefix(upper, "NOT ") {
		sub := strings.TrimSpace(expr[3:])
		cl, code, ok := parseClause(sub)
		if !ok {
			return "", "", false
		}
		return Not(cl), "b.Not(" + code + ")", true
	}
	if left, right, ok := splitComparison(expr, "="); ok {
		return buildEqualityClause(left, right)
	}
	return Clause(expr), fmt.Sprintf("b.Clause(%q)", expr), false
}

func buildEqualityClause(left, right string) (Clause, string, bool) {
	left = strings.TrimSpace(left)
	right = strings.TrimSpace(right)
	leftIsColumn := isColumnReference(left)
	rightIsColumn := isColumnReference(right)

	switch {
	case leftIsColumn && !rightIsColumn:
		rhsExp, rhsCode, ok := operandToBuilderExp(right, true)
		if !ok {
			return Clause(""), "", false
		}
		return Eq(left, rhsExp), fmt.Sprintf("b.Eq(%q, %s)", left, rhsCode), true
	case !leftIsColumn && rightIsColumn:
		lhsExp, lhsCode, ok := operandToBuilderExp(left, true)
		if !ok {
			return Clause(""), "", false
		}
		return Eq(right, lhsExp), fmt.Sprintf("b.Eq(%q, %s)", right, lhsCode), true
	case leftIsColumn && rightIsColumn:
		rhsExp, rhsCode, ok := operandToBuilderExp(right, true)
		if !ok {
			return Clause(""), "", false
		}
		return Eq(left, rhsExp), fmt.Sprintf("b.Eq(%q, %s)", left, rhsCode), true
	default:
		return Clause(""), "", false
	}
}

func operandToBuilderExp(operand string, allowIdentifier bool) (Exp, string, bool) {
	op := strings.TrimSpace(operand)
	if op == "" {
		return Exp(""), "", false
	}
	if strings.HasPrefix(op, "'") && strings.HasSuffix(op, "'") {
		value := strings.ReplaceAll(op[1:len(op)-1], "''", "'")
		return String(value), fmt.Sprintf("b.String(%q)", value), true
	}
	upper := strings.ToUpper(op)
	if upper == "TRUE" {
		return Bool(true), "b.Bool(true)", true
	}
	if upper == "FALSE" {
		return Bool(false), "b.Bool(false)", true
	}
	if i, err := strconv.ParseInt(op, 10, 64); err == nil {
		return Int64(i), fmt.Sprintf("b.Int64(%d)", i), true
	}
	if allowIdentifier && isColumnReference(op) {
		return Ident(op), fmt.Sprintf("b.Ident(%q)", op), true
	}
	switch strings.ToLower(op) {
	case "auth.uid()":
		return AuthUID(), "b.AuthUID()", true
	case "now()":
		return Now(), "b.Now()", true
	}
	if strings.Contains(op, "(") {
		name, args := parseFunctionCall(op)
		switch name {
		case "auth.uid":
			return AuthUID(), "b.AuthUID()", true
		case "current_setting":
			if len(args) == 1 && args[0].kind == argString {
				return CurrentSetting(args[0].value), fmt.Sprintf("b.CurrentSetting(%q)", args[0].value), true
			}
			if len(args) == 2 && args[0].kind == argString && args[1].kind == argIdentifier {
				return CurrentSettingAs(args[0].value, args[1].value), fmt.Sprintf("b.CurrentSettingAs(%q, %q)", args[0].value, args[1].value), true
			}
		}
	}
	return Raw(op), fmt.Sprintf("b.Raw(%q)", op), true
}

func removeQualifiers(expr string, qualifiers ...ClauseQualifier) string {
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

func splitByLogical(expr, operator string) []string {
	trimmed := strings.TrimSpace(expr)
	if trimmed == "" {
		return nil
	}
	upper := strings.ToUpper(trimmed)
	op := " " + operator + " "
	var parts []string
	depth := 0
	inString := false
	start := 0
	for i := 0; i < len(trimmed); i++ {
		ch := trimmed[i]
		if ch == '\'' {
			if inString && i+1 < len(trimmed) && trimmed[i+1] == '\'' {
				i++
				continue
			}
			inString = !inString
			continue
		}
		if inString {
			continue
		}
		switch ch {
		case '(':
			depth++
		case ')':
			depth--
		}
		if depth == 0 && i+len(op) <= len(trimmed) && upper[i:i+len(op)] == op {
			part := strings.TrimSpace(trimmed[start:i])
			if part != "" {
				parts = append(parts, part)
			}
			i += len(op) - 1
			start = i + 1
		}
	}
	if len(parts) == 0 {
		return nil
	}
	last := strings.TrimSpace(trimmed[start:])
	if last != "" {
		parts = append(parts, last)
	}
	return parts
}

func splitComparison(expr, operator string) (string, string, bool) {
	depth := 0
	inString := false
	for i := 0; i < len(expr); i++ {
		ch := expr[i]
		if ch == '\'' {
			if inString && i+1 < len(expr) && expr[i+1] == '\'' {
				i++
				continue
			}
			inString = !inString
			continue
		}
		if inString {
			continue
		}
		switch ch {
		case '(':
			depth++
		case ')':
			depth--
		}
		if depth == 0 && ch == '=' {
			left := strings.TrimSpace(expr[:i])
			right := strings.TrimSpace(expr[i+1:])
			return left, right, true
		}
	}
	return "", "", false
}

func isColumnReference(expr string) bool {
	if expr == "" {
		return false
	}
	for _, r := range expr {
		if !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '.') {
			return false
		}
	}
	return !strings.Contains(expr, "..") && !strings.HasSuffix(expr, ".")
}

type parsedArg struct {
	kind  argKind
	value string
}

type argKind int

const (
	argUnknown argKind = iota
	argString
	argIdentifier
)

func parseFunctionCall(expr string) (string, []parsedArg) {
	nameEnd := strings.Index(expr, "(")
	if nameEnd <= 0 || !strings.HasSuffix(expr, ")") {
		return "", nil
	}
	name := strings.TrimSpace(expr[:nameEnd])
	argsStr := strings.TrimSpace(expr[nameEnd+1 : len(expr)-1])
	if argsStr == "" {
		return name, nil
	}
	args := splitFunctionArgs(argsStr)
	parsed := make([]parsedArg, 0, len(args))
	for _, a := range args {
		a = strings.TrimSpace(a)
		if a == "" {
			continue
		}
		if strings.HasPrefix(a, "'") && strings.HasSuffix(a, "'") {
			parsed = append(parsed, parsedArg{kind: argString, value: strings.ReplaceAll(a[1:len(a)-1], "''", "'")})
			continue
		}
		if isColumnReference(a) {
			parsed = append(parsed, parsedArg{kind: argIdentifier, value: a})
			continue
		}
		parsed = append(parsed, parsedArg{kind: argUnknown, value: a})
	}
	return name, parsed
}

func splitFunctionArgs(args string) []string {
	depth := 0
	inString := false
	start := 0
	var parts []string
	for i := 0; i < len(args); i++ {
		ch := args[i]
		if ch == '\'' {
			if inString && i+1 < len(args) && args[i+1] == '\'' {
				i++
				continue
			}
			inString = !inString
			continue
		}
		if inString {
			continue
		}
		switch ch {
		case '(':
			depth++
		case ')':
			depth--
		case ',':
			if depth == 0 {
				parts = append(parts, strings.TrimSpace(args[start:i]))
				start = i + 1
			}
		}
	}
	parts = append(parts, strings.TrimSpace(args[start:]))
	return parts
}

func enclosedByOuterParentheses(expr string) bool {
	if len(expr) < 2 {
		return false
	}
	if expr[0] != '(' || expr[len(expr)-1] != ')' {
		return false
	}
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
