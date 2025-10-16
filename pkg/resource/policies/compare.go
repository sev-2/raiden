package policies

import (
	"strings"

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

	if source.Definition != target.Definition {
		updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdatePolicyDefinition)
	}

	if (source.Check == nil && target.Check != nil) || (source.Check != nil && target.Check == nil) || (source.Check != nil && target.Check != nil && *source.Check != *target.Check) {
		updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdatePolicyCheck)
	}

	if !stringsEqualUnordered(source.Roles, target.Roles) {
		updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdatePolicyRoles)
	}

	diffResult.IsConflict = len(updateItem.ChangeItems) > 0
	diffResult.DiffItems = updateItem
	return
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
