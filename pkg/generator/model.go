package generator

import (
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/hashicorp/go-hclog"
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/postgres"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/utils"
)

var ModelLogger hclog.Logger = logger.HcLog().Named("generator.model")

// ----- Define type, variable and constant -----
type (
	Rls struct {
		CanWrite []string
		CanRead  []string
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
		Relations  []state.Relation
		RlsTag     string
		RlsEnable  bool
		RlsForced  bool
		StructName string
		Schema     string
	}

	GenerateModelInput struct {
		Table     objects.Table
		Relations []state.Relation
		Policies  objects.Policies
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
	raiden.ModelBase
{{- range .Columns }}
	{{ .Name | ToGoIdentifier }} {{ .Type }} ` + "`{{ .Tag }}`" + `
{{- end }}

	// Table information
	Metadata string ` + "`json:\"-\" schema:\"{{ .Schema}}\" rlsEnable:\"{{ .RlsEnable }}\" rlsForced:\"{{ .RlsForced }}\"`" + `

	// Access control
	Acl string ` + "`json:\"-\" {{ .RlsTag }}`" + `
	
{{- if gt (len .Relations) 0 }}

	// Relations
{{- end }}
{{- range .Relations }}
	{{ .Table | ToGoIdentifier }} {{ .Type }} ` + "`{{ .Tag }}`" + `
{{- end }}
}
`
)

func GenerateModels(basePath string, tables []*GenerateModelInput, generateFn GenerateFn) (err error) {
	folderPath := filepath.Join(basePath, ModelDir)
	ModelLogger.Trace("create models folder if not exist", "path", folderPath)
	if exist := utils.IsFolderExists(folderPath); !exist {
		if err := utils.CreateFolder(folderPath); err != nil {
			return err
		}
	}

	for i := range tables {
		t := tables[i]
		if err := GenerateModel(folderPath, t, generateFn); err != nil {
			return err
		}
	}

	return nil
}

func GenerateModel(folderPath string, input *GenerateModelInput, generateFn GenerateFn) error {
	// define binding func
	funcMaps := []template.FuncMap{
		{"ToGoIdentifier": utils.SnakeCaseToPascalCase},
		{"ToSnakeCase": utils.ToSnakeCase},
	}

	// map column data
	columns, importsPath := MapTableAttributes(input.Table)
	rlsTag := BuildRlsTag(input.Policies)
	raidenPath := "github.com/sev-2/raiden"
	importsPath = append(importsPath, raidenPath)

	// define file path
	filePath := filepath.Join(folderPath, fmt.Sprintf("%s.%s", input.Table.Name, "go"))

	// build relation tag

	mapRelationName := make(map[string]bool)
	relation := make([]state.Relation, 0)

	for i := range input.Relations {
		r := input.Relations[i]

		if r.RelationType == raiden.RelationTypeManyToMany {
			key := fmt.Sprintf("%s_%s", input.Table.Name, r.Table)
			_, exist := mapRelationName[key]
			if exist {
				r.Table = fmt.Sprintf("%ss", r.Through)
			} else {
				mapRelationName[key] = true
			}
		}

		r.Tag = BuildJoinTag(&r)
		relation = append(relation, r)
	}

	// set data
	data := GenerateModelData{
		Package:    "models",
		Imports:    importsPath,
		StructName: utils.SnakeCaseToPascalCase(input.Table.Name),
		Columns:    columns,
		Schema:     input.Table.Schema,
		RlsTag:     rlsTag,
		RlsEnable:  input.Table.RLSEnabled,
		RlsForced:  input.Table.RLSForced,
		Relations:  relation,
	}

	// setup generate input param
	generateInput := GenerateInput{
		BindData:     data,
		FuncMap:      funcMaps,
		Template:     ModelTemplate,
		TemplateName: "modelTemplate",
		OutputPath:   filePath,
	}

	ModelLogger.Debug("generate model", "path", generateInput.OutputPath)
	return generateFn(generateInput, nil)
}

// map table to column, map pg type to go type and get dependency import path
func MapTableAttributes(table objects.Table) (columns []GenerateModelColumn, importsPath []string) {
	importsMap := make(map[string]any)
	mapPrimaryKey := map[string]bool{}
	for _, k := range table.PrimaryKeys {
		mapPrimaryKey[k.Name] = true
	}

	for _, c := range table.Columns {
		column := GenerateModelColumn{
			Name: c.Name,
			Tag:  buildColumnTag(c, mapPrimaryKey),
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

func buildColumnTag(c objects.Column, mapPk map[string]bool) string {
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

func BuildRlsTag(rlsList objects.Policies) string {
	var rls Rls

	var readUsingTag, writeCheckTag, writeUsingTag string
	for _, v := range rlsList {
		switch v.Command {
		case objects.PolicyCommandSelect:
			if v.Name == getPolicyName(objects.PolicyCommandSelect, v.Table) {
				rls.CanRead = append(rls.CanRead, v.Roles...)
				if v.Definition != "" {
					readUsingTag = v.Definition
				}
			}
		case objects.PolicyCommandInsert, objects.PolicyCommandUpdate, objects.PolicyCommandDelete:
			if v.Name == getPolicyName(objects.PolicyCommandInsert, v.Table) {
				if len(rls.CanWrite) == 0 {
					rls.CanWrite = append(rls.CanWrite, v.Roles...)
				}

				if len(writeCheckTag) == 0 && v.Check != nil {
					writeCheckTag = *v.Check
				}
			}

			if v.Name == getPolicyName(objects.PolicyCommandUpdate, v.Table) && len(rls.CanWrite) == 0 {
				if len(rls.CanWrite) == 0 {
					rls.CanWrite = append(rls.CanWrite, v.Roles...)
				}

				if len(writeCheckTag) == 0 && v.Check != nil {
					writeCheckTag = *v.Check
				}

				if len(writeUsingTag) == 0 && v.Definition != "" {
					writeUsingTag = v.Definition
				}
			}

			if v.Name == getPolicyName(objects.PolicyCommandDelete, v.Table) && len(rls.CanWrite) == 0 {
				if len(rls.CanWrite) == 0 {
					rls.CanWrite = append(rls.CanWrite, v.Roles...)
				}

				if len(writeUsingTag) == 0 && v.Definition != "" {
					writeUsingTag = v.Definition
				}
			}
		}
	}

	rlsTag := fmt.Sprintf("read:%q write:%q", strings.Join(rls.CanRead, ","), strings.Join(rls.CanWrite, ","))
	if len(readUsingTag) > 0 {
		cleanTag := strings.TrimLeft(strings.TrimRight(readUsingTag, ")"), "(")
		rlsTag = fmt.Sprintf("%s readUsing:%q", rlsTag, cleanTag)
	}

	if len(writeCheckTag) > 0 {
		cleanTag := strings.TrimLeft(strings.TrimRight(writeCheckTag, ")"), "(")
		rlsTag = fmt.Sprintf("%s writeCheck:%q", rlsTag, cleanTag)
	}

	if len(writeUsingTag) > 0 {
		cleanTag := strings.TrimLeft(strings.TrimRight(writeUsingTag, ")"), "(")
		rlsTag = fmt.Sprintf("%s writeUsing:%q", rlsTag, cleanTag)
	}

	return rlsTag
}

func BuildJoinTag(r *state.Relation) string {
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

func getPolicyName(command objects.PolicyCommand, tableName string) string {
	return strings.ToLower(fmt.Sprintf("enable %s access for table %s", command, tableName))
}
