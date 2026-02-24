package policies

import (
	"strings"

	"github.com/sev-2/raiden/pkg/builder"
	"github.com/sev-2/raiden/pkg/supabase/objects"
)

func Compare(sourcePolicies, targetPolicies []objects.Policy) error {
	diffResult := CompareList(sourcePolicies, targetPolicies)
	if len(diffResult) > 0 {
		return PrintDiffResult(diffResult)
	}
	return nil
}

type CompareDiffResult struct {
	Name           string
	SourceResource objects.Policy
	TargetResource objects.Policy
	DiffItems      objects.UpdatePolicyParam
	IsConflict     bool
}

func CompareList(sourcePolicies, targetPolicies []objects.Policy) (diffResult []CompareDiffResult) {
	mapTargetPolicies := make(map[string]objects.Policy)
	for i := range targetPolicies {
		r := targetPolicies[i]
		mapTargetPolicies[comparePolicyKey(r)] = r
	}
	for i := range sourcePolicies {
		p := sourcePolicies[i]

		tp, isExist := mapTargetPolicies[comparePolicyKey(p)]
		if !isExist {
			continue
		}
		diffResult = append(diffResult, CompareItem(p, tp))
	}

	return
}

// comparePolicyKey builds a unique key for matching policies. Policies with the
// same name on different tables are distinct; keying by name alone causes
// cross-table mismatches.
func comparePolicyKey(p objects.Policy) string {
	sch := strings.ToLower(p.Schema)
	table := strings.ToLower(p.Table)
	name := strings.ToLower(p.Name)
	if sch == "" && table == "" {
		return name
	}
	return sch + "." + table + "." + name
}

func CompareItem(source, target objects.Policy) (diffResult CompareDiffResult) {
	updateItem := objects.UpdatePolicyParam{
		Name:       target.Name,
		OldSchema:  target.Schema,
		OldTable:   target.Table,
		OldAction:  target.Action,
		OldCommand: target.Command,
		OldRoles:   append([]string(nil), target.Roles...),
	}

	// assign diff result object
	diffResult.SourceResource = source
	diffResult.TargetResource = target
	sourceName := strings.ToLower(source.Name)
	targetName := strings.ToLower(target.Name)
	if sourceName != targetName {
		updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdatePolicyName)
	}

	if !strings.EqualFold(source.Schema, target.Schema) {
		updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdatePolicySchema)
	}

	if !strings.EqualFold(source.Table, target.Table) {
		updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdatePolicyTable)
	}

	if !strings.EqualFold(source.Action, target.Action) {
		updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdatePolicyAction)
	}

	if !strings.EqualFold(string(source.Command), string(target.Command)) {
		updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdatePolicyCommand)
	}

	if shouldCompareDefinition(source.Command, target.Command) {
		qualifiers := []builder.ClauseQualifier{
			{Schema: source.Schema, Table: source.Table},
			{Schema: target.Schema, Table: target.Table},
		}
		sourceDefinition := normalizeComparableClause(source.Definition, qualifiers...)
		targetDefinition := normalizeComparableClause(target.Definition, qualifiers...)
		if sourceDefinition != targetDefinition && !equivalentBucketClauses(sourceDefinition, targetDefinition) {
			updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdatePolicyDefinition)
		}
	}

	if shouldCompareCheck(source.Command, target.Command) {
		qualifiers := []builder.ClauseQualifier{
			{Schema: source.Schema, Table: source.Table},
			{Schema: target.Schema, Table: target.Table},
		}
		sourceCheck, hasSource := normalizeComparableOptionalClause(source.Check, qualifiers...)
		targetCheck, hasTarget := normalizeComparableOptionalClause(target.Check, qualifiers...)
		if hasSource != hasTarget || (hasSource && sourceCheck != targetCheck && !equivalentBucketClauses(sourceCheck, targetCheck)) {
			updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdatePolicyCheck)
		}
	}

	if !stringsEqualUnordered(source.Roles, target.Roles) {
		updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdatePolicyRoles)
	}

	diffResult.IsConflict = len(updateItem.ChangeItems) > 0
	diffResult.DiffItems = updateItem
	return
}

func shouldCompareDefinition(sourceCmd, targetCmd objects.PolicyCommand) bool {
	return commandUsesDefinition(sourceCmd) || commandUsesDefinition(targetCmd)
}

func shouldCompareCheck(sourceCmd, targetCmd objects.PolicyCommand) bool {
	return commandUsesCheck(sourceCmd) || commandUsesCheck(targetCmd)
}

func normalizeComparableClause(sql string, qualifiers ...builder.ClauseQualifier) string {
	normalized := builder.NormalizeClauseSQL(sql, qualifiers...)
	normalized = normalizeBooleanLiteral(normalized)
	normalized = uppercaseLogicalKeywords(normalized)
	normalized = strings.ReplaceAll(normalized, "(TRUE)", "TRUE")
	normalized = strings.ReplaceAll(normalized, "(FALSE)", "FALSE")
	return collapseWhitespace(strings.TrimSpace(normalized))
}

func normalizeComparableOptionalClause(clause *string, qualifiers ...builder.ClauseQualifier) (string, bool) {
	if clause == nil {
		return "", false
	}
	trimmed := strings.TrimSpace(*clause)
	if trimmed == "" {
		return "", false
	}
	return normalizeComparableClause(trimmed, qualifiers...), true
}

func normalizeBooleanLiteral(input string) string {
	switch strings.ToUpper(strings.TrimSpace(input)) {
	case "":
		return ""
	case "TRUE":
		return "TRUE"
	case "FALSE":
		return "FALSE"
	default:
		return strings.TrimSpace(input)
	}
}

func collapseWhitespace(input string) string {
	return strings.Join(strings.Fields(input), " ")
}

func uppercaseLogicalKeywords(input string) string {
	tokens := strings.Fields(input)
	for i := range tokens {
		leading, core, trailing := splitToken(tokens[i])
		switch strings.ToUpper(core) {
		case "AND", "OR", "NOT":
			core = strings.ToUpper(core)
		}
		tokens[i] = leading + core + trailing
	}
	return strings.Join(tokens, " ")
}

func splitToken(token string) (leading, core, trailing string) {
	leading = leadingParens(token)
	trailing = trailingParens(token)
	core = token[len(leading) : len(token)-len(trailing)]
	return leading, core, trailing
}

func leadingParens(token string) string {
	i := 0
	for i < len(token) && (token[i] == '(' || token[i] == '[') {
		i++
	}
	return token[:i]
}

func trailingParens(token string) string {
	i := len(token)
	for i > 0 && (token[i-1] == ')' || token[i-1] == ']') {
		i--
	}
	return token[i:]
}

func commandUsesDefinition(cmd objects.PolicyCommand) bool {
	switch normalizePolicyCommand(cmd) {
	case string(objects.PolicyCommandInsert):
		return false
	default:
		return true
	}
}

func commandUsesCheck(cmd objects.PolicyCommand) bool {
	switch normalizePolicyCommand(cmd) {
	case string(objects.PolicyCommandSelect), string(objects.PolicyCommandDelete):
		return false
	default:
		return true
	}
}

func normalizePolicyCommand(cmd objects.PolicyCommand) string {
	return strings.ToUpper(strings.TrimSpace(string(cmd)))
}

func equivalentBucketClauses(a, b string) bool {
	if !strings.Contains(strings.ToLower(a), "bucket_id") || !strings.Contains(strings.ToLower(b), "bucket_id") {
		return false
	}
	normalize := func(input string) string {
		s := strings.ToLower(input)
		s = strings.ReplaceAll(s, "\"", "")
		s = strings.ReplaceAll(s, "(", "")
		s = strings.ReplaceAll(s, ")", "")
		s = strings.ReplaceAll(s, " ", "")
		s = strings.ReplaceAll(s, "=true", "=true")
		s = strings.ReplaceAll(s, "true=", "true=")
		return s
	}
	return normalize(a) == normalize(b)
}

func stringsEqualUnordered(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	seen := make(map[string]int, len(a))
	for _, v := range a {
		seen[v]++
	}
	for _, v := range b {
		count, ok := seen[v]
		if !ok {
			return false
		}
		if count == 1 {
			delete(seen, v)
		} else {
			seen[v] = count - 1
		}
	}
	return len(seen) == 0
}
