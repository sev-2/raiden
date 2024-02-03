package generator

import (
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/postgres"
	"github.com/sev-2/raiden/pkg/supabase"
	"github.com/sev-2/raiden/pkg/utils"
)

// ----- Define type, van and constant -----
type Rls struct {
	CanWrite []string
	CanRead  []string
}

type Relation struct {
	Table        string
	Type         string
	SourceColumn string
	TargetColumn string
}

type GenerateModelData struct {
	Columns    []map[string]any
	Imports    []string
	Package    string
	Relations  []Relation
	RlsTag     string
	StructName string
	Schema     string
}

const (
	ModelDir      = "internal/models"
	ModelTemplate = `package {{ .Package }}
{{- if gt (len .Imports) 0 }}

import (
{{- range .Imports}}
	"{{.}}"
{{- end}}
)
{{- end }}

type {{ .StructName }} struct {
{{- range .Columns }}
	{{ .Name | ToGoIdentifier }} {{ .GoType }} ` + "`json:\"{{ .Name | ToSnakeCase }},omitempty\" column:\"{{ .Name | ToSnakeCase }}\"`" + `
{{- end }}

	// Table information
	Metadata string ` + "`json:\"metadata,omitempty\" schema:\"{{ .Schema}}\"`" + `

	// Access control
	Acl string ` + "`json:\"acl,omitempty\" {{ .RlsTag }}`" + `
	
{{- if gt (len .Relations) 0 }}

	// Relations
{{- end }}
{{- range .Relations }}
	{{ .Table | ToGoIdentifier }} {{ .Type }} ` + "`json:\"{{ .Table | ToSnakeCase }},omitempty\" sourceColumn:\"{{ .SourceColumn }}\" targetColumn:\"{{ .TargetColumn }}\"`" + `
{{- end }}
}
`
)

func GenerateModels(basePath string, tables []supabase.Table, rlsList supabase.Policies, generateFn GenerateFn) (err error) {
	folderPath := filepath.Join(basePath, ModelDir)
	logger.Debugf("GenerateModels - create %s folder if not exist", folderPath)
	if exist := utils.IsFolderExists(folderPath); !exist {
		if err := utils.CreateFolder(folderPath); err != nil {
			return err
		}
	}

	for i, v := range tables {
		searchTable := tables[i].Name
		GenerateModel(folderPath, v, rlsList.FilterByTable(searchTable), generateFn)
	}

	return nil
}

func GenerateModel(folderPath string, table supabase.Table, rlsList supabase.Policies, generateFn GenerateFn) error {
	// define binding func
	funcMaps := []template.FuncMap{
		{"ToGoIdentifier": utils.SnakeCaseToPascalCase},
		{"ToGoType": postgres.ToGoType},
		{"ToSnakeCase": utils.ToSnakeCase},
	}

	// map column data
	columns, importsPath := mapTableAttributes(table)
	rlsTag := generateRlsTag(rlsList)

	// map relations
	relations := mapTableRelation(table)

	// define file path
	filePath := filepath.Join(folderPath, fmt.Sprintf("%s.%s", table.Name, "go"))

	// set data
	data := GenerateModelData{
		Package:    "models",
		Imports:    importsPath,
		StructName: utils.SnakeCaseToPascalCase(table.Name),
		Columns:    columns,
		Schema:     table.Schema,
		RlsTag:     rlsTag,
		Relations:  relations,
	}

	// setup generate input param
	input := GenerateInput{
		BindData:     data,
		FuncMap:      funcMaps,
		Template:     ModelTemplate,
		TemplateName: "modelTemplate",
		OutputPath:   filePath,
	}

	logger.Debugf("GenerateModels - generate model to %s", input.OutputPath)
	return generateFn(input)
}

// map table to column, map pg type to go type and get dependency import path
func mapTableAttributes(table supabase.Table) (columns []map[string]any, importsPath []string) {
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

func mapTableRelation(table supabase.Table) (relations []Relation) {
	for i := range table.Relationships {
		r := table.Relationships[i]
		typePrefix := "*"

		if r.SourceTableName != table.Name {
			foundSourceColumn := false
			for i := range table.Columns {
				c := table.Columns[i]
				if c.Name == r.SourceColumnName {
					foundSourceColumn = true
					break
				}
			}

			if !foundSourceColumn {
				typePrefix = "[]*"
			}

			relations = append(relations, Relation{
				Table:        r.SourceTableName,
				Type:         typePrefix + utils.SnakeCaseToPascalCase(r.SourceTableName),
				SourceColumn: r.SourceColumnName,
				TargetColumn: r.TargetColumnName,
			})
			continue
		}

		relations = append(relations, Relation{
			Table:        r.TargetTableName,
			Type:         typePrefix + utils.SnakeCaseToPascalCase(r.TargetTableName),
			SourceColumn: r.TargetColumnName,
			TargetColumn: r.SourceColumnName,
		})
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
