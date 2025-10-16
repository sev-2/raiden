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
		mapTargetPolicies[strings.ToLower(r.Name)] = r
	}
	for i := range sourcePolicies {
		p := sourcePolicies[i]

		tp, isExist := mapTargetPolicies[strings.ToLower(p.Name)]
		if !isExist {
			continue
		}
		diffResult = append(diffResult, CompareItem(p, tp))
	}

	return
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
		sourceDefinition := builder.NormalizeClauseSQL(source.Definition, qualifiers...)
		targetDefinition := builder.NormalizeClauseSQL(target.Definition, qualifiers...)
		if sourceDefinition != targetDefinition {
			updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdatePolicyDefinition)
		}
	}

	if shouldCompareCheck(source.Command, target.Command) {
		qualifiers := []builder.ClauseQualifier{
			{Schema: source.Schema, Table: source.Table},
			{Schema: target.Schema, Table: target.Table},
		}
		sourceCheck, hasSource := normalizeOptionalClause(source.Check, qualifiers...)
		targetCheck, hasTarget := normalizeOptionalClause(target.Check, qualifiers...)
		if hasSource != hasTarget || (hasSource && sourceCheck != targetCheck) {
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

func normalizeOptionalClause(clause *string, qualifiers ...builder.ClauseQualifier) (string, bool) {
	if clause == nil {
		return "", false
	}
	trimmed := strings.TrimSpace(*clause)
	if trimmed == "" {
		return "", false
	}
	return builder.NormalizeClauseSQL(trimmed, qualifiers...), true
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
