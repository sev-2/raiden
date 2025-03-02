package types

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"text/template"

	"github.com/fatih/color"
	"github.com/sev-2/raiden/pkg/resource/migrator"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/utils"
)

// ----- print diff section -----
type DiffType string

const (
	DiffTypeCreate DiffType = "create"
	DiffTypeUpdate DiffType = "update"
	DiffTypeDelete DiffType = "delete"
)

func PrintDiffResult(diffResult []CompareDiffResult) error {
	if len(diffResult) == 0 {
		return nil
	}

	isConflict := false
	for i := range diffResult {
		d := diffResult[i]
		if d.IsConflict {
			PrintDiff(d)
			if !isConflict {
				isConflict = true
			}
		}
	}

	if isConflict {
		return errors.New("canceled import process, you have conflict in type. please fix it first")
	}

	return nil
}

func PrintDiff(diffData CompareDiffResult) {
	if len(diffData.DiffItems.ChangeItems) == 0 {
		return
	}
	fileName := utils.ToSnakeCase(diffData.TargetResource.Name)
	printScope := color.New(color.FgHiBlack).PrintfFunc()

	changes := make([]string, 0)
	for _, v := range diffData.DiffItems.ChangeItems {
		switch v {
		case objects.UpdateTypeName:
			var tgName, scName string
			if diffData.SourceResource.Name != "" {
				scName = diffData.SourceResource.Name
			}

			if diffData.TargetResource.Name != "" {
				tgName = diffData.TargetResource.Name
			}

			diffStr, err := GenerateDiffMessage(fileName, DiffTypeUpdate, v, tgName, scName)
			if err != nil {
				Logger.Error("print diff type name error", "msg", err.Error())
				continue
			}
			changes = append(changes, diffStr)
		// case objects.UpdateTypeIsReplication:
		// case objects.UpdateTypeIsSuperUser:
		case objects.UpdateTypeSchema:
			var tgName, scName string
			if diffData.SourceResource.Schema != "" {
				scName = diffData.SourceResource.Schema
			}
			if diffData.TargetResource.Schema != "" {
				tgName = diffData.TargetResource.Schema
			}

			diffStr, err := GenerateDiffMessage(fileName, DiffTypeUpdate, v, tgName, scName)
			if err != nil {
				Logger.Error("print diff type schema error", "msg", err.Error())
				continue
			}
			changes = append(changes, diffStr)
		case objects.UpdateTypeFormat:
			var tgName, scName string
			if diffData.SourceResource.Format != "" {
				scName = diffData.SourceResource.Format
			}
			if diffData.TargetResource.Format != "" {
				tgName = diffData.TargetResource.Format
			}

			diffStr, err := GenerateDiffMessage(fileName, DiffTypeUpdate, v, tgName, scName)
			if err != nil {
				Logger.Error("print diff types error", "msg", err.Error())
				continue
			}
			changes = append(changes, diffStr)

		case objects.UpdateTypeComment:
			var diffType DiffType
			var value, changedValue string

			if diffData.TargetResource.Comment == nil && diffData.SourceResource.Comment != nil && *diffData.SourceResource.Comment != "" {
				diffType = DiffTypeCreate
				value = *diffData.SourceResource.Comment
			}

			if diffData.TargetResource.Comment != nil && *diffData.TargetResource.Comment != "" && diffData.SourceResource.Comment == nil {
				diffType = DiffTypeDelete
				value = *diffData.SourceResource.Comment
			}

			if diffData.TargetResource.Comment != nil && diffData.SourceResource.Comment != nil && *diffData.TargetResource.Comment != *diffData.SourceResource.Comment {
				diffType = DiffTypeUpdate
				value = *diffData.TargetResource.Comment
				changedValue = *diffData.SourceResource.Comment
			}

			diffStr, err := GenerateDiffMessage(fileName, diffType, v, value, changedValue)
			if err != nil {
				Logger.Error("print diff types error", "msg", err.Error())
				continue
			}
			changes = append(changes, diffStr)

		}
	}

	printScope("*** Found diff in %s/%s.go ***\n", "/internal/types", fileName)
	fmt.Println(strings.Join(changes, ""))
	printScope("*** End found diff ***\n")
}

// ----- generate message section ------
const DiffTemplate = ` 
 {{- if or (eq .Type "create") (eq .Type "delete")}}
 {{ .Symbol }} func (r *{{ .Name  | ToGoIdentifier }}) %s {
 %s
 {{ .Symbol }} }
 {{- end}}
 {{- if eq .Type "update"}}
 func (r *{{ .Name | ToGoIdentifier }}) %s {
   %s
 }
{{- end}}
  `

const FuncBodyTemplate = "{{ .Symbol }} return {{ .Value }}"
const FuncBodyUpdateTemplate = "{{ .Symbol }} return {{ .Value }}  >>> {{ .ChangeValue }}"

func buildDiffTemplate(funcDecl string, bodyTemplate string, updateBodyTemplate string) string {
	if bodyTemplate == "" {
		bodyTemplate = FuncBodyTemplate
	}

	if updateBodyTemplate == "" {
		updateBodyTemplate = FuncBodyUpdateTemplate
	}

	return fmt.Sprintf(DiffTemplate, funcDecl, bodyTemplate, funcDecl, updateBodyTemplate)
}

func getDiffSymbol(diffType DiffType) string {
	printAdd := color.New(color.FgHiGreen).SprintfFunc()
	printRemove := color.New(color.FgHiRed).SprintfFunc()
	printUpdate := color.New(color.FgHiYellow).SprintfFunc()

	var symbol string
	switch diffType {
	case DiffTypeCreate:
		symbol = printAdd("+")
	case DiffTypeUpdate:
		symbol = printUpdate("~")
	case DiffTypeDelete:
		symbol = printRemove("-")
	}
	return symbol
}

func GenerateDiffMessage(name string, diffType DiffType, updateType objects.UpdateDataType, value string, changeValue string) (string, error) {
	param := map[string]any{
		"Name":        name,
		"Type":        diffType,
		"Value":       value,
		"ChangeValue": changeValue,
		"Symbol":      getDiffSymbol(diffType),
	}

	tmplStr := ""
	switch updateType {
	case objects.UpdateTypeName:
		tmplStr = buildDiffTemplate("Name() string", "", "")
	case objects.UpdateTypeSchema:
		tmplStr = buildDiffTemplate("Schema() *string", "", "")
	case objects.UpdateTypeFormat:
		tmplStr = buildDiffTemplate("Format() string", "", "")
	case objects.UpdateTypeAttributes:
		tmplStr = buildDiffTemplate("Attributes() []string", "", "")
	case objects.UpdateTypeEnums:
		tmplStr = buildDiffTemplate("Enums() []string", "", "")
	case objects.UpdateTypeComment:
		tmplStr = buildDiffTemplate("Format() *string", "", "")
	default:
		return "", errors.New("unsupported update type")
	}

	funcMaps := []template.FuncMap{
		{"ToGoIdentifier": utils.SnakeCaseToPascalCase},
	}

	tmplInstance := template.New("generate diff")
	for _, tm := range funcMaps {
		tmplInstance.Funcs(tm)
	}

	tmpl, err := tmplInstance.Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("error parsing : %v", err)
	}

	var buff bytes.Buffer
	if err := tmpl.Execute(&buff, param); err != nil {
		return "", err
	}

	return buff.String(), nil
}

// ----- diff change -----

func GetDiffChangeMessage(items []MigrateItem) string {
	newData := []string{}
	deleteData := []string{}
	updateData := []string{}

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
			newData = append(newData, fmt.Sprintf("- %s", name))
		case migrator.MigrateTypeUpdate:
			diffMessage, err := GenerateDiffChangeUpdateMessage(name, item)
			if err != nil {
				Logger.Error("print change type error", "msg", err.Error())
				continue
			}
			updateData = append(updateData, diffMessage)
		case migrator.MigrateTypeDelete:
			deleteData = append(deleteData, fmt.Sprintf("- %s", name))
		}
	}

	changeMsg, err := GenerateDiffChangeMessage(newData, updateData, deleteData)
	if err != nil {
		Logger.Error("print change type error", "msg", err.Error())
		return ""
	}
	return changeMsg
}

const DiffChangeTemplate = `
  {{- if gt (len .NewData) 0}}
  New Type
  {{- range .NewData}}
  {{.}}
  {{- end }}
  {{- end -}}
  {{- if gt (len .UpdateData) 0}}
  Update Type
  {{- range .UpdateData}}
  {{.}}
  {{- end }}
  {{- end -}}
  {{- if gt (len .DeleteData) 0}}
  Delete Type
  {{- range .DeleteData}}
  {{.}}
  {{- end }}
  {{- end -}}
  `

func GenerateDiffChangeMessage(newData []string, updateData []string, deleteData []string) (string, error) {
	param := map[string]any{
		"NewData":    newData,
		"UpdateData": updateData,
		"DeleteData": deleteData,
	}

	tmplInstance := template.New("generate diff change type")
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

const DiffChangeUpdateTemplate = `  - Update Role {{ .Name }}
  {{- if gt (len .ChangeItems) 0}}
      Change Configuration
      {{- range .ChangeItems}}
      {{.}}
      {{- end }}
  {{- end -}}
  `

func GenerateDiffChangeUpdateMessage(name string, item MigrateItem) (string, error) {
	diffItems := item.MigrationItems

	var changeMsgArr []string
	for i := range diffItems.ChangeItems {
		c := diffItems.ChangeItems[i]
		switch c {
		case objects.UpdateTypeName:
			changeMsgArr = append(changeMsgArr, fmt.Sprintf("- %s : %v >>> %v", "name", item.OldData.Name, item.NewData.Name))
		case objects.UpdateTypeSchema:
			changeMsgArr = append(changeMsgArr, fmt.Sprintf("- %s : %v >>> %v", "schema", item.OldData.Schema, item.NewData.Schema))
		case objects.UpdateTypeComment:
			changeMsgArr = append(changeMsgArr, fmt.Sprintf("- %s : %v >>> %v", "comment", item.OldData.Comment, item.NewData.Comment))
		case objects.UpdateTypeFormat:
			changeMsgArr = append(changeMsgArr, fmt.Sprintf("- %s : %v >>> %v", "format", item.OldData.Format, item.NewData.Format))
		case objects.UpdateTypeEnums:
			oldData := strings.Join(item.OldData.Enums, ",")
			newData := strings.Join(item.NewData.Enums, ",")
			changeMsgArr = append(changeMsgArr, fmt.Sprintf("- %s : %v >>> %v", "enums", oldData, newData))
		}
	}

	param := map[string]any{
		"Name":        name,
		"ChangeItems": changeMsgArr,
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
