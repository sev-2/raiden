package tables

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"text/template"

	"github.com/fatih/color"
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/generator"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/utils"
)

func PrintDiffResult(diffResult []CompareDiffResult, sRelation MapRelations, tRelation MapRelations) error {
	isConflict := false
	for i := range diffResult {
		d := diffResult[i]
		if d.IsConflict {
			PrintDiff(d, sRelation, tRelation)
			if !isConflict {
				isConflict = true
			}
		}
	}

	if isConflict {
		return errors.New("canceled import process, you have conflict table. please fix it first")
	}

	return nil
}

func PrintDiff(diffData CompareDiffResult, sRelation MapRelations, tRelation MapRelations) {
	fileName := utils.ToSnakeCase(diffData.TargetResource.Name)
	printScope := color.New(color.FgHiBlack).PrintfFunc()

	diffMessage, err := GenerateDiffMessage(diffData, sRelation, tRelation)
	if err != nil {
		Logger.Error("print diff rpc error", "msg", err.Error())
		return
	}

	printScope("*** Found diff in %s/%s.go ***\n", "/internal/models", fileName)
	fmt.Println(diffMessage)
	printScope("*** End found diff ***\n")
}

// ----- generate message section ------
const DiffTemplate = ` 
{{- .Columns}}
{{- .Metadata}}
{{- .Acl}}
{{- .Relations}}
  `

func GenerateDiffMessage(diffData CompareDiffResult, sRelation MapRelations, tRelation MapRelations) (string, error) {
	printUpdate := color.New(color.FgHiYellow).SprintfFunc()
	printIndent := color.New(color.FgHiBlack).SprintfFunc()
	symbol := printUpdate("~")
	fromIndent := printIndent("from:")
	toIndent := printIndent("to:")

	var diffColumnStr, diffMetadata, diffAclStr, diffRelationStr string

	// start generate message
	if len(diffData.DiffItems.ChangeItems) > 0 {
		tMetadata := fmt.Sprintf(
			"%s %s Metadata string `json:\"-\" schema:\"%s\" rlsEnable:\"%t\" rlsForced:\"%t\"`",
			symbol, fromIndent, diffData.TargetResource.Schema,
			diffData.TargetResource.RLSEnabled, diffData.TargetResource.RLSForced,
		)

		sMetadata := fmt.Sprintf(
			"%s %s Metadata string `json:\"-\" schema:\"%s\" rlsEnable:\"%t\" rlsForced:\"%t\"`",
			symbol, toIndent, diffData.SourceResource.Schema,
			diffData.SourceResource.RLSEnabled, diffData.SourceResource.RLSForced,
		)

		diffMetadata = fmt.Sprintf("\n%s\n%s", tMetadata, sMetadata)
	}

	if len(diffData.DiffItems.ChangeColumnItems) > 0 {
		diffColumns := []string{}

		// mas source column
		mapSColumns, _ := generator.MapTableAttributes(diffData.SourceResource)
		mapTColumns, _ := generator.MapTableAttributes(diffData.TargetResource)

		// find source column
		for ic := range diffData.DiffItems.ChangeColumnItems {
			var foundSColumn, foundTColumn generator.GenerateModelColumn
			changeColumn := diffData.DiffItems.ChangeColumnItems[ic]

			for fi := range mapSColumns {
				c := mapSColumns[fi]
				if changeColumn.Name == c.Name {
					foundSColumn = c
					break
				}
			}

			for fi := range mapTColumns {
				c := mapTColumns[fi]
				if changeColumn.Name == c.Name {
					foundTColumn = c
					break
				}
			}

			if foundSColumn.Name == "" && foundTColumn.Name == "" {
				continue
			}

			tColumn := fmt.Sprintf(
				"%s %s %s %s `%s`",
				symbol, fromIndent,
				utils.SnakeCaseToPascalCase(foundTColumn.Name),
				foundTColumn.Type,
				foundTColumn.Tag,
			)

			sColumn := fmt.Sprintf(
				"%s %s %s %s `%s`",
				symbol, toIndent,
				utils.SnakeCaseToPascalCase(foundSColumn.Name),
				foundSColumn.Type,
				foundSColumn.Tag,
			)

			diffColumns = append(diffColumns, fmt.Sprintf("%s\n%s", tColumn, sColumn))
		}

		if len(diffColumns) > 0 {
			diffColumnStr = fmt.Sprintf("\n%s\n", strings.Join(diffColumns, "\n"))
		}
	}

	if len(diffData.DiffItems.ChangeRelationItems) > 0 {
		var sFoundRelations, tFoundRelations []*state.Relation
		var sRelationStr, tRelationStr string

		sKey := fmt.Sprintf("%s.%s", diffData.SourceResource.Schema, diffData.SourceResource.Name)
		if r, exist := sRelation[sKey]; exist {
			sFoundRelations = r
		}

		tKey := fmt.Sprintf("%s.%s", diffData.TargetResource.Schema, diffData.TargetResource.Name)
		if r, exist := tRelation[tKey]; exist {
			tFoundRelations = r
		}

		if len(sFoundRelations) == 0 || sFoundRelations == nil {
			sRelationStr = "not implemented"
		} else {
			mapRelationName := make(map[string]bool)
			relations := make([]state.Relation, 0)
			for i := range sFoundRelations {
				r := sFoundRelations[i]
				if r == nil {
					continue
				}

				if r.RelationType == raiden.RelationTypeManyToMany {
					key := fmt.Sprintf("%s_%s", diffData.SourceResource.Name, r.Table)
					_, exist := mapRelationName[key]
					if exist {
						r.Table = fmt.Sprintf("%ss", r.Through)
					} else {
						mapRelationName[key] = true
					}
				}

				r.Tag = generator.BuildJoinTag(r)
				relations = append(relations, *r)
			}

			if len(relations) > 0 {
				for i := range relations {
					r := relations[i]
					sRelationStr += fmt.Sprintf(
						"%s *%s `json:\"%s,omitempty\" %s\n",
						symbol, utils.SnakeCaseToPascalCase(r.Table),
						r.Type, r.Tag,
					)
				}
			}
		}

		if len(tFoundRelations) == 0 || tFoundRelations == nil {
			sRelationStr = "not implemented"
		} else {
			mapRelationName := make(map[string]bool)
			relations := make([]state.Relation, 0)
			for i := range tFoundRelations {
				r := tFoundRelations[i]
				if r == nil {
					continue
				}

				if r.RelationType == raiden.RelationTypeManyToMany {
					key := fmt.Sprintf("%s_%s", diffData.TargetResource.Name, r.Table)
					_, exist := mapRelationName[key]
					if exist {
						r.Table = fmt.Sprintf("%ss", r.Through)
					} else {
						mapRelationName[key] = true
					}
				}

				r.Tag = generator.BuildJoinTag(r)
				relations = append(relations, *r)
			}

			if len(relations) > 0 {
				for i := range relations {
					r := relations[i]
					tRelationStr += fmt.Sprintf(
						"%s *%s `json:\"%s,omitempty\" %s\n",
						symbol, utils.SnakeCaseToPascalCase(r.Table),
						r.Type, r.Tag,
					)
				}
			}
		}

		diffRelationStr = fmt.Sprintf("%s %s \n%s\n%s %s \n%s", symbol, fromIndent, tRelationStr, symbol, toIndent, sRelationStr)
	}

	param := map[string]any{
		"Columns":  diffColumnStr,
		"Metadata": diffMetadata,
		// Handle ACL Compare
		"Acl":       diffAclStr,
		"Relations": diffRelationStr,
	}

	tmplInstance := template.New("generate diff")
	tmpl, err := tmplInstance.Parse(DiffTemplate)
	if err != nil {
		return "", fmt.Errorf("error parsing : %v", err)
	}

	var buff bytes.Buffer
	if err := tmpl.Execute(&buff, param); err != nil {
		return "", err
	}

	return buff.String(), nil
}
