package generator

import (
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/postgres"
	"github.com/sev-2/raiden/pkg/supabase"
	"github.com/sev-2/raiden/pkg/utils"
)

// ----- Define type, variable and constant -----
type (
	Rls struct {
		CanWrite []string
		CanRead  []string
	}

	Relation struct {
		Table        string
		Type         string
		RelationType raiden.RelationType
		PrimaryKey   string
		ForeignKey   string
		Tag          string
		*JoinRelation
	}

	JoinRelation struct {
		SourcePrimaryKey      string
		JoinsSourceForeignKey string

		TargetPrimaryKey     string
		JoinTargetForeignKey string

		Through string
	}

	GenerateModelColumn struct {
		Name string
		Type string
		Tag  string
	}

	GenerateModelData struct {
		Columns    []GenerateModelColumn
		Imports    []string
		Package    string
		Relations  []Relation
		RlsTag     string
		StructName string
		Schema     string
	}

	GenerateModelInput struct {
		Table     supabase.Table
		Relations []Relation
	}
)

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
	{{ .Name | ToGoIdentifier }} {{ .Type }} ` + "`{{ .Tag }}`" + `
{{- end }}

	// Table information
	Metadata string ` + "`json:\"metadata,omitempty\" schema:\"{{ .Schema}}\"`" + `

	// Access control
	Acl string ` + "`json:\"acl,omitempty\" {{ .RlsTag }}`" + `
	
{{- if gt (len .Relations) 0 }}

	// Relations
{{- end }}
{{- range .Relations }}
	{{ .Table | ToGoIdentifier }} {{ .Type }} ` + "`{{ .Tag }}`" + `
{{- end }}
}
`
)

func GenerateModels(basePath string, tables []*GenerateModelInput, rlsList supabase.Policies, generateFn GenerateFn) (err error) {
	folderPath := filepath.Join(basePath, ModelDir)
	logger.Debugf("GenerateModels - create %s folder if not exist", folderPath)
	if exist := utils.IsFolderExists(folderPath); !exist {
		if err := utils.CreateFolder(folderPath); err != nil {
			return err
		}
	}

	for i, v := range tables {
		searchTable := tables[i].Table.Name
		GenerateModel(folderPath, v, rlsList.FilterByTable(searchTable), generateFn)
	}

	return nil
}

func GenerateModel(folderPath string, input *GenerateModelInput, rlsList supabase.Policies, generateFn GenerateFn) error {
	// define binding func
	funcMaps := []template.FuncMap{
		{"ToGoIdentifier": utils.SnakeCaseToPascalCase},
		{"ToSnakeCase": utils.ToSnakeCase},
	}

	// map column data
	columns, importsPath := mapTableAttributes(input.Table)
	rlsTag := generateRlsTag(rlsList)

	// define file path
	filePath := filepath.Join(folderPath, fmt.Sprintf("%s.%s", input.Table.Name, "go"))

	// set data
	data := GenerateModelData{
		Package:    "models",
		Imports:    importsPath,
		StructName: utils.SnakeCaseToPascalCase(input.Table.Name),
		Columns:    columns,
		Schema:     input.Table.Schema,
		RlsTag:     rlsTag,
		Relations:  input.Relations,
	}

	// setup generate input param
	generateInput := GenerateInput{
		BindData:     data,
		FuncMap:      funcMaps,
		Template:     ModelTemplate,
		TemplateName: "modelTemplate",
		OutputPath:   filePath,
	}

	logger.Debugf("GenerateModels - generate model to %s", generateInput.OutputPath)
	return generateFn(generateInput)
}

// map table to column, map pg type to go type and get dependency import path
func mapTableAttributes(table supabase.Table) (columns []GenerateModelColumn, importsPath []string) {
	importsMap := make(map[string]any)
	mapPrimaryKey := map[string]bool{}
	for _, k := range table.PrimaryKeys {
		mapPrimaryKey[k.Name] = true
	}

	for _, c := range table.Columns {
		column := GenerateModelColumn{
			Name: c.Name,
			Tag:  generateColumnTag(c, mapPrimaryKey),
			Type: postgres.ToGoType(postgres.DataType(c.DataType), c.IsNullable),
		}

		splitType := strings.Split(column.Type, ".")
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

func generateColumnTag(c supabase.Column, mapPk map[string]bool) string {
	var tags []string

	// append json tag
	jsonTag := fmt.Sprintf("json:%q", utils.ToSnakeCase(c.Name)+",omitempty")
	tags = append(tags, jsonTag)

	// append column tag
	columnTags := []string{
		fmt.Sprintf("name:%s", c.Name),
	}

	if postgres.IsValidDataType(c.DataType) {
		pdType := postgres.GetPgDataTypeName(postgres.DataType(c.DataType), true)
		columnTags = append(columnTags, "type:"+string(pdType))
	}

	_, exist := mapPk[c.Name]
	if exist {
		columnTags = append(columnTags, "primaryKey")
	}

	if c.IdentityGeneration != nil {
		if identityStr, isString := c.IdentityGeneration.(string); isString && len(identityStr) > 0 {
			columnTags = append(columnTags, "autoIncrement")
		}
	}

	if c.IsNullable {
		columnTags = append(columnTags, "nullable")
	} else {
		columnTags = append(columnTags, "nullable:false")
	}

	if c.DefaultValue != "" {
		defaultStr, isString := c.DefaultValue.(string)
		if isString {
			columnTags = append(columnTags, "default:"+defaultStr)
		}
	}

	if c.IsUnique {
		columnTags = append(columnTags, "unique")
	}

	tags = append(tags, fmt.Sprintf("column:%q", strings.Join(columnTags, ";")))

	return strings.Join(tags, " ")
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

// Relation
func (r *Relation) BuildTag() string {
	var tags []string
	var joinTags []string

	// append json tag
	jsonTag := fmt.Sprintf("json:%q", utils.ToSnakeCase(r.Table)+",omitempty")
	tags = append(tags, jsonTag)

	// append relation type tag
	relTypeTag := fmt.Sprintf("joinType:%s", r.RelationType)
	joinTags = append(joinTags, relTypeTag)

	// append PK tag
	if r.PrimaryKey != "" {
		pk := fmt.Sprintf("primaryKey:%s", r.PrimaryKey)
		joinTags = append(joinTags, pk)
	}

	// append FK tag
	if r.PrimaryKey != "" {
		fk := fmt.Sprintf("foreignKey:%s", r.ForeignKey)
		joinTags = append(joinTags, fk)
	}

	if r.RelationType == raiden.RelationTypeManyToMany && r.JoinRelation != nil {
		th := fmt.Sprintf("through:%s", r.Through)
		joinTags = append(joinTags, th)

		// append source primary key
		spk := fmt.Sprintf("sourcePrimaryKey:%s", r.SourcePrimaryKey)
		joinTags = append(joinTags, spk)

		// append join source foreign key
		jspk := fmt.Sprintf("sourceForeignKey:%s", r.JoinsSourceForeignKey)
		joinTags = append(joinTags, jspk)

		// append target primary key
		tpk := fmt.Sprintf("targetPrimaryKey:%s", r.TargetPrimaryKey)
		joinTags = append(joinTags, tpk)

		// append join target foreign key
		jtpk := fmt.Sprintf("targetForeign:%s", r.JoinsSourceForeignKey)
		joinTags = append(joinTags, jtpk)
	}
	tags = append(tags, fmt.Sprintf("join:%q", strings.Join(joinTags, ";")))

	return strings.Join(tags, " ")
}
