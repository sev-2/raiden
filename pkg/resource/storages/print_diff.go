package storages

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"text/template"

	"github.com/fatih/color"
	"github.com/sev-2/raiden/pkg/generator"
	"github.com/sev-2/raiden/pkg/resource/migrator"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/utils"
)

// ----- Print diff section -----
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
		return errors.New("canceled import process, you have conflict in storage. please fix it first")
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
		case objects.UpdateBucketIsPublic:
			var scIsPublic = "false"
			if diffData.SourceResource.Public {
				scIsPublic = "true"
			}

			var tgIsPublic = "false"
			if diffData.TargetResource.Public {
				tgIsPublic = "true"
			}

			diffStr, err := GenerateDiffMessage(fileName, DiffTypeUpdate, v, tgIsPublic, scIsPublic)
			if err != nil {
				Logger.Error("print diff storage error", "msg", err.Error())
				continue
			}
			changes = append(changes, diffStr)
		case objects.UpdateBucketAllowedMimeTypes:
			var diffType DiffType
			var value, changedValue string

			if len(diffData.TargetResource.AllowedMimeTypes) == 0 && len(diffData.SourceResource.AllowedMimeTypes) > 0 {
				diffType = DiffTypeCreate
				value = generator.GenerateArrayDeclaration(reflect.ValueOf(diffData.SourceResource.AllowedMimeTypes), false)
			} else if len(diffData.TargetResource.AllowedMimeTypes) > 0 && len(diffData.SourceResource.AllowedMimeTypes) == 0 {
				diffType = DiffTypeDelete
				value = generator.GenerateArrayDeclaration(reflect.ValueOf(diffData.SourceResource.AllowedMimeTypes), false)
			} else {
				diffType = DiffTypeUpdate
				value = generator.GenerateArrayDeclaration(reflect.ValueOf(diffData.TargetResource.AllowedMimeTypes), false)
				changedValue = generator.GenerateArrayDeclaration(reflect.ValueOf(diffData.TargetResource.AllowedMimeTypes), false)
			}

			diffStr, err := GenerateDiffMessage(fileName, diffType, v, value, changedValue)
			if err != nil {
				Logger.Error("print diff storage error", "msg", err.Error())
				continue
			}
			changes = append(changes, diffStr)
		case objects.UpdateBucketFileSizeLimit:
			var diffType DiffType
			var value, changedValue string

			if diffData.TargetResource.FileSizeLimit == nil && diffData.SourceResource.FileSizeLimit != nil && *diffData.SourceResource.FileSizeLimit > 0 {
				diffType = DiffTypeCreate
				value = strconv.Itoa(*diffData.SourceResource.FileSizeLimit)
			}

			if diffData.TargetResource.FileSizeLimit != nil && *diffData.TargetResource.FileSizeLimit > 0 && diffData.SourceResource.FileSizeLimit == nil {
				diffType = DiffTypeDelete
				value = strconv.Itoa(*diffData.SourceResource.FileSizeLimit)
			}

			if diffData.TargetResource.FileSizeLimit != nil && diffData.SourceResource.FileSizeLimit != nil && *diffData.TargetResource.FileSizeLimit != *diffData.SourceResource.FileSizeLimit {
				diffType = DiffTypeUpdate
				value = strconv.Itoa(*diffData.TargetResource.FileSizeLimit)
				changedValue = strconv.Itoa(*diffData.SourceResource.FileSizeLimit)
			}

			diffStr, err := GenerateDiffMessage(fileName, diffType, v, value, changedValue)
			if err != nil {
				Logger.Error("print diff storage error", "msg", err.Error())
				continue
			}
			changes = append(changes, diffStr)
		}
	}

	printScope("*** Found diff in %s/%s.go ***\n", "/internal/storages", fileName)
	fmt.Println(strings.Join(changes, ""))
	printScope("*** End found diff ***\n")
}

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
				Logger.Error("print change storage error", "msg", err.Error())
				continue
			}
			updateData = append(updateData, diffMessage)
		case migrator.MigrateTypeDelete:
			deleteData = append(deleteData, fmt.Sprintf("- %s", name))
		}
	}

	changeMsg, err := GenerateDiffChangeMessage(newData, updateData, deleteData)
	if err != nil {
		Logger.Error("print change storage error", "msg", err.Error())
		return ""
	}
	return changeMsg
}

// ----- generate message section -----
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

func GenerateDiffMessage(storageName string, diffType DiffType, updateType objects.UpdateBucketType, value string, changeValue string) (string, error) {
	param := map[string]any{
		"Name":        storageName,
		"Type":        diffType,
		"Value":       value,
		"ChangeValue": changeValue,
		"Symbol":      getDiffSymbol(diffType),
	}

	tmplStr := ""
	switch updateType {
	case objects.UpdateBucketIsPublic:
		tmplStr = buildDiffTemplate("Public() bool", "", "")
	case objects.UpdateBucketFileSizeLimit:
		tmplStr = buildDiffTemplate("FileSizeLimit() int", "", "")
	case objects.UpdateBucketAllowedMimeTypes:
		tmplStr = buildDiffTemplate("AllowedMimeTypes() []string", "", "")
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
const DiffChangeTemplate = `
  {{- if gt (len .NewData) 0}}
  New Storage
  {{- range .NewData}}
  {{.}}
  {{- end }}
  {{- end -}}
  {{- if gt (len .UpdateData) 0}}
  Update Storage
  {{- range .UpdateData}}
  {{.}}
  {{- end }}
  {{- end -}}
  {{- if gt (len .DeleteData) 0}}
  Delete Storage
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

	tmplInstance := template.New("generate diff change storage")
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

const DiffChangeUpdateTemplate = `  - Update Storage {{ .Name }}
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
		case objects.UpdateBucketIsPublic:
			changeMsgArr = append(changeMsgArr, fmt.Sprintf("- %s : %t >>> %t", "is public", item.OldData.Public, item.NewData.Public))
		case objects.UpdateBucketFileSizeLimit:
			changeMsgArr = append(changeMsgArr, fmt.Sprintf("- %s : %v >>> %v", "file size limit", item.OldData.FileSizeLimit, item.NewData.FileSizeLimit))
		case objects.UpdateBucketAllowedMimeTypes:
			changeMsgArr = append(changeMsgArr, fmt.Sprintf("- %s : %s >>> %s", "allowed mime type", strings.Join(item.OldData.AllowedMimeTypes, ","), strings.Join(item.NewData.AllowedMimeTypes, ",")))
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
