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
	"github.com/sev-2/raiden/pkg/resource/migrator"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase/objects"
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
		Logger.Error("print diff table error", "msg", err.Error())
		return
	}

	printScope("*** Found diff in %s/%s.go ***\n", "/internal/models", fileName)
	fmt.Println(diffMessage)
	printScope("*** End found diff ***\n")
}

func GetDiffChangeMessage(items []MigrateItem) string {
	newTable := []string{}
	deleteTable := []string{}
	updateTable := []string{}

	for i := range items {
		item := items[i]

		var name string
		if item.NewData.Name != "" {
			name = item.NewData.Name
		} else if item.OldData.Name != "" {
			name = item.OldData.Name
		}

		switch item.Type {
		case migrator.MigrateTypeCreate:
			newTable = append(newTable, fmt.Sprintf("- %s", name))
		case migrator.MigrateTypeUpdate:
			// if Logger.GetLevel() == hclog.Trace {
			diffMessage, err := GenerateDiffChangeUpdateMessage(name, item)
			if err != nil {
				Logger.Error("print change table error", "msg", err.Error())
				continue
			}
			updateTable = append(updateTable, diffMessage)
			// } else {
			// 	updateTable = append(updateTable, fmt.Sprintf("- %s", name))
			// }
		case migrator.MigrateTypeDelete:
			deleteTable = append(deleteTable, fmt.Sprintf("- %s", name))
		}
	}

	changeMsg, err := GenerateDiffChangeMessage(newTable, updateTable, deleteTable)
	if err != nil {
		Logger.Error("print change table error", "msg", err.Error())
		return ""
	}
	return changeMsg
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
		var sRelationArr, tRelationArr []string

		sKey := fmt.Sprintf("%s.%s", diffData.SourceResource.Schema, diffData.SourceResource.Name)
		if r, exist := sRelation[sKey]; exist {
			sFoundRelations = r
		}

		tKey := fmt.Sprintf("%s.%s", diffData.TargetResource.Schema, diffData.TargetResource.Name)
		if r, exist := tRelation[tKey]; exist {
			tFoundRelations = r
		}

		mapChangedByTargetTable := map[string][]objects.UpdateRelationItem{}
		for i := range diffData.DiffItems.ChangeRelationItems {
			dItem := diffData.DiffItems.ChangeRelationItems[i]
			changeTables, exist := mapChangedByTargetTable[dItem.Data.TargetTableName]
			if exist {
				changeTables = append(changeTables, dItem)
				mapChangedByTargetTable[dItem.Data.TargetTableName] = changeTables
			} else {
				mapChangedByTargetTable[dItem.Data.TargetTableName] = []objects.UpdateRelationItem{dItem}
			}
		}

		// normalize sRelations
		if len(sFoundRelations) > 0 {
			mapRelationName := make(map[string]bool)
			relations := make([]*state.Relation, 0)
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
				relations = append(relations, r)
			}
			sFoundRelations = relations
		}

		// normalize tRelations
		if len(tFoundRelations) > 0 {
			mapRelationName := make(map[string]bool)
			relations := make([]*state.Relation, 0)
			for i := range tFoundRelations {
				r := tFoundRelations[i]
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
				relations = append(relations, r)
			}
			tFoundRelations = relations
		}

		for k, diffItems := range mapChangedByTargetTable {
			var fSource, fTarget *state.Relation

			// find source
			for i := range diffItems {
				di := diffItems[i]
				findKey := fmt.Sprintf("%s_%s_%s", k, di.Data.SourceColumnName, di.Data.TargetColumnName)

				for si := range sFoundRelations {
					if fSource != nil {
						break
					}

					sRelation := sFoundRelations[si]
					if sRelation == nil {
						continue
					}

					sKey := fmt.Sprintf(
						"%s_%s_%s",
						sRelation.Table,
						sRelation.ForeignKey,
						sRelation.PrimaryKey,
					)

					if findKey == sKey {
						fSource = sRelation
					}
				}

				for ti := range tFoundRelations {
					if fTarget != nil {
						break
					}

					tRelation := tFoundRelations[ti]
					if tRelation == nil {
						continue
					}
					tKey := fmt.Sprintf("%s_%s_%s", tRelation.Table, tRelation.ForeignKey, tRelation.PrimaryKey)

					if findKey == tKey {
						fTarget = tRelation
					}
				}
			}

			if fSource == nil && fTarget == nil {
				continue
			}

			if fSource != nil {
				sRelationArr = append(sRelationArr, fmt.Sprintf(
					"%s *%s `json:\"%s,omitempty\" %s",
					symbol, utils.SnakeCaseToPascalCase(fSource.Table),
					fSource.Type, fSource.Tag,
				))
			} else {
				sRelationArr = append(sRelationArr, fmt.Sprintf(
					"%s not implemented", symbol,
				))
			}

			if fTarget != nil {
				tRelationArr = append(tRelationArr, fmt.Sprintf(
					"%s *%s `json:\"%s,omitempty\" %s",
					symbol, utils.SnakeCaseToPascalCase(fTarget.Table),
					fTarget.Type, fTarget.Tag,
				))
			} else {
				tRelationArr = append(tRelationArr, fmt.Sprintf(
					"%s not implemented", symbol,
				))
			}
		}

		diffRelationStr = fmt.Sprintf(
			"\n%s %s \n%s\n%s %s \n%s",
			symbol, fromIndent, strings.Join(tRelationArr, "\n"),
			symbol, toIndent, strings.Join(sRelationArr, "\n"),
		)
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

const DiffChangeTemplate = `
  {{- if gt (len .NewTable) 0}}
  New Table
  {{- range .NewTable}}
  {{.}}
  {{- end }}
  {{- end -}}
  {{- if gt (len .UpdateTable) 0}}
  Update Table
  {{- range .UpdateTable}}
  {{.}}
  {{- end }}
  {{- end -}}
  {{- if gt (len .DeleteTable) 0}}
  Delete Table
  {{- range .DeleteTable}}
  {{.}}
  {{- end }}
  {{- end -}}
  `

const DiffChangeUpdateTemplate = `  - Update Table {{ .Name }}
  {{- if gt (len .ChangeItems) 0}}
      Change Configuration
      {{- range .ChangeItems}}
      {{.}}
      {{- end }}
  {{- end -}}
  {{- if gt (len .ChangeColumns) 0}}
      Change Columns
      {{- range .ChangeColumns}}
      {{.}}
      {{- end }}
  {{- end -}}
  {{- if gt (len .ChangeRelations) 0}}
      Change Relations
      {{- range .ChangeRelations}}
      {{.}}
      {{- end }}
  {{- end -}}
  `

func GenerateDiffChangeMessage(newTable []string, updateTable []string, deleteTable []string) (string, error) {
	param := map[string]any{
		"NewTable":    newTable,
		"UpdateTable": updateTable,
		"DeleteTable": deleteTable,
	}

	tmplInstance := template.New("generate diff change")
	tmpl, err := tmplInstance.Parse(DiffChangeTemplate)
	if err != nil {
		return "", fmt.Errorf("error parsing : %v", err)
	}

	var buff bytes.Buffer
	if err := tmpl.Execute(&buff, param); err != nil {
		return "", err
	}

	return buff.String(), nil
}

func GenerateDiffChangeUpdateMessage(name string, item MigrateItem) (string, error) {
	diffItems := item.MigrationItems

	var changeMsgArr, changeColumnMsgArr, changeRelationArr []string
	for i := range diffItems.ChangeItems {
		c := diffItems.ChangeItems[i]
		switch c {
		case objects.UpdateTableSchema:
			changeMsgArr = append(changeMsgArr, fmt.Sprintf("- %s : %s >>> %s", "schema", item.OldData.Schema, item.NewData.Schema))
		case objects.UpdateTableName:
			changeMsgArr = append(changeMsgArr, fmt.Sprintf("- %s : %s >>> %s", "name", item.OldData.Name, item.NewData.Name))
		case objects.UpdateTableRlsEnable:
			changeMsgArr = append(changeMsgArr, fmt.Sprintf("- %s : %t >>> %t", "rls enable", item.OldData.RLSEnabled, item.NewData.RLSEnabled))
		case objects.UpdateTableRlsForced:
			changeMsgArr = append(changeMsgArr, fmt.Sprintf("- %s : %t >>> %t", "rls forced", item.OldData.RLSEnabled, item.NewData.RLSEnabled))
		case objects.UpdateTableReplicaIdentity:
			changeMsgArr = append(changeMsgArr, fmt.Sprintf("- %s : %s >>> %s", "replica identity", item.OldData.ReplicaIdentity, item.NewData.ReplicaIdentity))
		}
	}

	for i := range diffItems.ChangeColumnItems {
		c := diffItems.ChangeColumnItems[i]

		var oldColumn, newColum objects.Column

		// find old column detail
		for ii := range item.OldData.Columns {
			oc := item.OldData.Columns[ii]
			if oc.Name == c.Name {
				oldColumn = oc
			}
		}

		// find new column detail
		for ii := range item.NewData.Columns {
			nc := item.NewData.Columns[ii]
			if nc.Name == c.Name {
				newColum = nc
			}
		}

		var updateItemArr []string
		for ic := range c.UpdateItems {
			cType := c.UpdateItems[ic]
			switch cType {
			case objects.UpdateColumnNew:
				updateItemArr = append(updateItemArr, fmt.Sprintf("- %s : %s", "create new column", newColum.Name))
			case objects.UpdateColumnDelete:
				updateItemArr = append(updateItemArr, fmt.Sprintf("- %s : %s", "delete column", oldColumn.Name))
			case objects.UpdateColumnName:
				updateItemArr = append(updateItemArr, fmt.Sprintf("- %s : %s >>> %s", "name", oldColumn.Name, newColum.Name))
			case objects.UpdateColumnDefaultValue:
				var oldValue, newValue = "nil", "nil"
				if oldColumn.DefaultValue != nil {
					oldValue = utils.ConvertAllToString(oldColumn.DefaultValue)
				}

				if newColum.DefaultValue != nil {
					newValue = utils.ConvertAllToString(newColum.DefaultValue)
				}
				updateItemArr = append(updateItemArr, fmt.Sprintf("- %s : %v >>> %v", "default value", oldValue, newValue))
			case objects.UpdateColumnDataType:
				updateItemArr = append(updateItemArr, fmt.Sprintf("- %s : %v >>> %v", "data type", oldColumn.DataType, newColum.DataType))
			case objects.UpdateColumnUnique:
				updateItemArr = append(updateItemArr, fmt.Sprintf("- %s : %t >>> %t", "is unique", oldColumn.IsUnique, newColum.IsUnique))
			case objects.UpdateColumnNullable:
				updateItemArr = append(updateItemArr, fmt.Sprintf("- %s : %t >>> %t", "is nullable", oldColumn.IsNullable, newColum.IsNullable))
			case objects.UpdateColumnIdentity:
				updateItemArr = append(updateItemArr, fmt.Sprintf("- %s : %t >>> %t", "is identity", oldColumn.IsIdentity, newColum.IsIdentity))
			}
		}

		var name = oldColumn.Name
		if name == "" {
			name = newColum.Name
		}

		updateMessageStr := fmt.Sprintf("- update column %q\n        %s", name, strings.Join(updateItemArr, "\n        "))
		changeColumnMsgArr = append(changeColumnMsgArr, updateMessageStr)
	}

	for i := range diffItems.ChangeRelationItems {
		c := diffItems.ChangeRelationItems[i]

		switch c.Type {
		case objects.UpdateRelationCreate:
			changeRelationArr = append(changeRelationArr, fmt.Sprintf("- %s : %s", "create new relation", c.Data.ConstraintName))
		case objects.UpdateRelationUpdate:
			changeRelationArr = append(changeRelationArr, fmt.Sprintf("- %s : %s", "update relation", c.Data.ConstraintName))
		case objects.UpdateRelationDelete:
			changeRelationArr = append(changeRelationArr, fmt.Sprintf("- %s : %s", "delete relation", c.Data.ConstraintName))
		}
	}

	param := map[string]any{
		"Name":            name,
		"ChangeItems":     changeMsgArr,
		"ChangeColumns":   changeColumnMsgArr,
		"ChangeRelations": changeRelationArr,
	}

	tmplInstance := template.New("generate diff change update")
	tmpl, err := tmplInstance.Parse(DiffChangeUpdateTemplate)
	if err != nil {
		return "", fmt.Errorf("error parsing : %v", err)
	}

	var buff bytes.Buffer
	if err := tmpl.Execute(&buff, param); err != nil {
		return "", err
	}

	return buff.String(), nil
}
