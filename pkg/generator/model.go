package generator

import (
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/sev-2/raiden/pkg/postgres"
	"github.com/sev-2/raiden/pkg/supabase"
	"github.com/sev-2/raiden/pkg/utils"
)

var modelDir = "models"
var modelTemplate = `package models
{{ if gt (len .Imports) 0 }}
import(
{{- range .Imports}}
	"{{.}}"
{{- end}}
)
{{ end }}
type {{ .StructName }} struct {
{{- range .Columns }}
	{{ .Name | ToGoIdentifier }} {{ .GoType }} ` + "`json:\"{{ .Name | ToSnakeCase }},omitempty\" column:\"{{ .Name | ToSnakeCase }}\"`" + `
{{- end }}

	Metadata string ` + "`schema:\"{{ .Schema}}\"`" + `
	Acl string ` + "`{{ .RlsTag }}`" + `
}
`

type Rls struct {
	CanWrite []string
	CanRead  []string
}

func GenerateModels(projectName string, tables []supabase.Table, rlsList supabase.Policies) (err error) {
	folderPath := filepath.Join(projectName, modelDir)
	err = utils.CreateFolder(folderPath)
	if err != nil {
		return err
	}

	for i, v := range tables {
		searchTable := tables[i].Name
		GenerateModel(projectName, v, rlsList.FilterByTable(searchTable))
	}

	return nil
}

func GenerateModel(projectName string, table supabase.Table, rlsList supabase.Policies) error {
	tmpl, err := template.New("modelTemplate").
		Funcs(template.FuncMap{"ToGoIdentifier": utils.SnakeCaseToPascalCase}).
		Funcs(template.FuncMap{"ToGoType": postgres.ToGoType}).
		Funcs(template.FuncMap{"ToSnakeCase": utils.ToSnakeCase}).
		Parse(modelTemplate)
	if err != nil {
		return fmt.Errorf("error parsing template : %v", err)
	}

	// Map column data
	columns, importsPath := mapTableToColumn(table)
	rlsTag := generateRlsTag(rlsList)

	// Create or open the output file
	folderPath := fmt.Sprintf("%s/%s", projectName, modelDir)
	file, err := createFile(getAbsolutePath(folderPath), table.Name, "go")
	if err != nil {
		return fmt.Errorf("failed create file %s : %v", table.Name, err)
	}
	defer file.Close()

	// Execute the template and write to the file
	err = tmpl.Execute(file, map[string]any{
		"Imports":    importsPath,
		"StructName": utils.SnakeCaseToPascalCase(table.Name),
		"Columns":    columns,
		"Schema":     table.Schema,
		"RlsTag":     rlsTag,
	})

	if err != nil {
		return fmt.Errorf("error executing template: %v", err)
	}

	return nil
}

// map table to column, map pg type to go type and get dependency import path
func mapTableToColumn(table supabase.Table) (columns []map[string]any, importsPath []string) {
	importsMap := make(map[string]any)
	columns = make([]map[string]any, 0)

	for _, c := range table.Columns {
		column := map[string]any{
			"Name": c.Name,
		}

		goType := postgres.ToGoType(c.DataType, c.IsNullable)
		column["GoType"] = goType

		splitType := strings.Split(goType, ".")
		if len(splitType) > 1 {
			importPackage := splitType[0]
			if c.IsNullable {
				importPackage = strings.TrimLeft(importPackage, "*")
			}

			var importPackageName string
			switch importPackage {
			case "time":
				importPackageName = importPackage
			case "uuid":
				importPackageName = "github.com/google/uuid"
			case "json":
				importPackageName = "encoding/json"
			}
			importsMap[importPackageName] = true
		}

		columns = append(columns, column)
	}

	for key := range importsMap {
		importsPath = append(importsPath, key)
	}

	return
}

func generateRlsTag(rlsList supabase.Policies) string {
	var rls Rls

	for _, v := range rlsList {
		switch v.Command {
		case supabase.PolicyCommandSelect:
			rls.CanWrite = append(rls.CanWrite, v.Roles...)
		case supabase.PolicyCommandInsert, supabase.PolicyCommandUpdate, supabase.PolicyCommandDelete:
			rls.CanWrite = append(rls.CanWrite, v.Roles...)
		}
	}

	rlsTag := fmt.Sprintf("read:%q write:%q", strings.Join(rls.CanRead, ","), strings.Join(rls.CanWrite, ","))
	return rlsTag
}
