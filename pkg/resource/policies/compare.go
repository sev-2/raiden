package policies

import (
	"strings"

	"github.com/sev-2/raiden/pkg/supabase/objects"
)

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
		mapTargetPolicies[r.Name] = r
	}
	for i := range sourcePolicies {
		p := sourcePolicies[i]

		tp, isExist := mapTargetPolicies[p.Name]
		if !isExist {
			continue
		}
		diffResult = append(diffResult, CompareItem(p, tp))
	}

	return
}

func CompareItem(source, target objects.Policy) (diffResult CompareDiffResult) {
	var updateItem objects.UpdatePolicyParam

	// assign diff result object
	diffResult.SourceResource = source
	diffResult.TargetResource = target
	updateItem.Name = source.Name

	sourceName := strings.ToLower(source.Name)
	targetName := strings.ToLower(target.Name)
	if sourceName != targetName {
		updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdatePolicyName)
	}

	if source.Definition != target.Definition {
		updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdatePolicyDefinition)
	}

	if (source.Check == nil && target.Check != nil) || (source.Check != nil && target.Check == nil) || (source.Check != nil && target.Check != nil && *source.Check != *target.Check) {
		updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdatePolicyCheck)
	}

	if len(source.Roles) != len(target.Roles) {
		updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdatePolicyRoles)
	} else {
		for sr := range source.Roles {
			isFound := false
			for tr := range target.Roles {
				if sr == tr {
					isFound = true
					break
				}
			}

			if !isFound {
				updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdatePolicyRoles)
				break
			}
		}
	}

	diffResult.IsConflict = len(updateItem.ChangeItems) > 0
	diffResult.DiffItems = updateItem
	return
}
