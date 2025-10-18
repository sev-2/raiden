package generator

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"unicode"

	"github.com/hashicorp/go-hclog"
	"github.com/jinzhu/inflection"
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/builder"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/postgres"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/utils"
)

var ModelLogger hclog.Logger = logger.HcLog().Named("generator.model")

// ----- Define type, variable and constant -----
type (
	GenerateModelColumn struct {
		Name string
		Type string
		Tag  string
	}

	GenerateModelData struct {
		Columns          []GenerateModelColumn
		Imports          []string
		Package          string
		Relations        []state.Relation
		RlsEnable        bool
		RlsForced        bool
		StructName       string
		Schema           string
		TableName        string
		Receiver         string
		HasConfigureAcl  bool
		ConfigureAclBody string
	}

	GenerateModelInput struct {
		Table          objects.Table
		Relations      []state.Relation
		Policies       objects.Policies
		ValidationTags state.ModelValidationTag
	}
)

const (
	ModelDir      = "internal/models"
	ModelTemplate = `package {{ .Package }}
{{- if gt (len .Imports) 0 }}

import (
{{- range .Imports}}
	{{.}}
{{- end}}
)
{{- end }}

type {{ .StructName }} struct {
	db.ModelBase

{{- range .Columns }}
	{{ .Name | ToGoIdentifier }} {{ .Type }} ` + "`{{ .Tag }}`" + `
{{- end }}

	// Table information
	Metadata string ` + "`json:\"-\" schema:\"{{ .Schema}}\" tableName:\"{{ .TableName }}\" rlsEnable:\"{{ .RlsEnable }}\" rlsForced:\"{{ .RlsForced }}\"`" + `

	// Access control
	Acl raiden.Acl

{{- if gt (len .Relations) 0 }}

	// Relations
{{- end }}
{{- range .Relations }}
	{{ .Table | ToGoIdentifier }} {{ .Type }} ` + "`{{ .Tag }}`" + `
{{- end }}
}

{{- if .HasConfigureAcl }}

func ({{ .Receiver }} *{{ .StructName }}) ConfigureAcl() {
{{ .ConfigureAclBody }}
}
{{- end }}
`
)

func GenerateModels(
	basePath string, projectName string, tables []*GenerateModelInput,
	mapDataType map[string]objects.Type, roleMap map[string]string, nativeRoleMap map[string]raiden.Role,
	generateFn GenerateFn,
) (err error) {
	folderPath := filepath.Join(basePath, ModelDir)
	ModelLogger.Trace("create models folder if not exist", "path", folderPath)
	if exist := utils.IsFolderExists(folderPath); !exist {
		if err := utils.CreateFolder(folderPath); err != nil {
			return err
		}
	}

	for i := range tables {
		t := tables[i]
		if err := GenerateModel(folderPath, projectName, t, mapDataType, roleMap, nativeRoleMap, generateFn); err != nil {
			return err
		}
	}

	return nil
}

func GenerateModel(folderPath string, projectName string, input *GenerateModelInput,
	mapDataType map[string]objects.Type, roleMap map[string]string, nativeRoleMap map[string]raiden.Role,
	generateFn GenerateFn,
) error {
	// define binding func
	funcMaps := []template.FuncMap{
		{"ToGoIdentifier": utils.SnakeCaseToPascalCase},
		{"ToSnakeCase": utils.ToSnakeCase},
	}

	// map column data
	columns, importsPath := MapTableAttributes(projectName, input.Table, mapDataType, input.ValidationTags)
	raidenPkgDbPath := "github.com/sev-2/raiden/pkg/db"
	importsPath = append(importsPath, raidenPkgDbPath)

	// define file path
	filePath := filepath.Join(folderPath, fmt.Sprintf("%s.%s", input.Table.Name, "go"))

	// build relation field
	relations := BuildRelationFields(input.Table, input.Relations)

	structName := utils.SnakeCaseToPascalCase(input.Table.Name)
	receiverName := strings.ToLower(string(structName[0]))
	moduleName := utils.ToGoModuleName(projectName)
	rolesImportPath := fmt.Sprintf("%s/internal/roles", moduleName)

	aclInfo, err := buildAclInfo(structName, receiverName, input.Table, input.Policies, roleMap, nativeRoleMap, nil)
	if err != nil {
		return err
	}

	importsPath = append(importsPath, "github.com/sev-2/raiden")
	if aclInfo.UseBuilder {
		importsPath = append(importsPath, `st "github.com/sev-2/raiden/pkg/builder"`)
	}
	if aclInfo.UseRoles {
		importsPath = append(importsPath, fmt.Sprintf("roles %q", rolesImportPath))
	}
	if aclInfo.UseNativeRoles {
		importsPath = append(importsPath, `native_role "github.com/sev-2/raiden/pkg/postgres/roles"`)
	}

	importsPath = normalizeImports(importsPath)

	// set data
	data := GenerateModelData{
		Package:          "models",
		Imports:          importsPath,
		StructName:       structName,
		Columns:          columns,
		Schema:           input.Table.Schema,
		TableName:        input.Table.Name,
		RlsEnable:        input.Table.RLSEnabled,
		RlsForced:        input.Table.RLSForced,
		Relations:        relations,
		Receiver:         receiverName,
		HasConfigureAcl:  aclInfo.HasConfigure,
		ConfigureAclBody: aclInfo.Body,
	}

	// setup generate input param
	generateInput := GenerateInput{
		BindData:     data,
		FuncMap:      funcMaps,
		Template:     ModelTemplate,
		TemplateName: "modelTemplate",
		OutputPath:   filePath,
	}

	// setup writer
	writer := FileWriter{FilePath: filePath}

	ModelLogger.Debug("generate model", "path", generateInput.OutputPath)
	return generateFn(generateInput, &writer)
}

// map table to column, map pg type to go type and get dependency import path
func MapTableAttributes(projectName string, table objects.Table, mapDataType map[string]objects.Type, validationTags state.ModelValidationTag) (columns []GenerateModelColumn, importsPath []string) {
	importsMap := make(map[string]any)
	mapPrimaryKey := map[string]bool{}
	for _, k := range table.PrimaryKeys {
		mapPrimaryKey[k.Name] = true
	}

	for _, c := range table.Columns {
		var userDataType *objects.Type

		// check data type
		if c.DataType == string(postgres.UserDefined) && mapDataType != nil {
			dataType, exist := mapDataType[c.Format]
			if exist {
				userDataType = &dataType
			}
		}

		column := GenerateModelColumn{
			Name: c.Name,
			Tag:  buildColumnTag(c, mapPrimaryKey, userDataType, validationTags),
		}

		if userDataType != nil {
			column.Type = fmt.Sprintf("types.%s", utils.SnakeCaseToPascalCase(userDataType.Name))
			typeImportPath := fmt.Sprintf("%s/internal/types", utils.ToGoModuleName(projectName))
			importsMap[typeImportPath] = true
		} else {
			column.Type = postgres.ToGoType(postgres.DataType(c.DataType), c.IsNullable)
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

			// custom type
			postgresCustomTypes := map[string]bool{
				"postgres.Point":     true,
				"*postgres.Point":    true,
				"postgres.DateTime":  true,
				"*postgres.DateTime": true,
				"postgres.Date":      true,
				"*postgres.Date":     true,
			}

			if postgresCustomTypes[column.Type] {
				importPackageName = "github.com/sev-2/raiden/pkg/postgres"
			}

			importsMap[importPackageName] = true
		}

		columns = append(columns, column)
	}

	for key := range importsMap {
		if key == "" {
			continue
		}
		importsPath = append(importsPath, key)
	}
	sort.Strings(importsPath)

	return
}

func resolvePolicyColumns(modelName, storageName string, tableMap map[string]objects.Table) (string, []objects.Column) {
	if modelName != "" {
		if table, ok := tableMap[modelName]; ok {
			return "p.model", table.Columns
		}
	}
	if storageName != "" {
		if table, ok := tableMap[storageName]; ok {
			return "p.storage", table.Columns
		}
	}
	return "", nil
}

func generateClauseCode(sql string, qualifier builder.ClauseQualifier, receiver string, columns []objects.Column) string {
	const alias = "st"
	trimmed := strings.TrimSpace(sql)
	if strings.EqualFold(trimmed, "TRUE") {
		return fmt.Sprintf("%s.True", alias)
	}
	if strings.EqualFold(trimmed, "FALSE") {
		return fmt.Sprintf("%s.False", alias)
	}

	_, clauseCode, ok := builder.UnmarshalClause(sql, qualifier)
	if !ok {
		normalized := strings.TrimSpace(builder.NormalizeClauseSQL(sql, qualifier))
		if strings.EqualFold(normalized, "TRUE") {
			return fmt.Sprintf("%s.True", alias)
		}
		if strings.EqualFold(normalized, "FALSE") {
			return fmt.Sprintf("%s.False", alias)
		}
		return fmt.Sprintf("%s.Clause(%q)", alias, normalized)
	}

	clauseCode = strings.ReplaceAll(clauseCode, "b.", alias+".")
	clauseCode = normalizeAliasBooleanClause(clauseCode, alias)
	if receiver != "" && len(columns) > 0 {
		return injectColumnReferences(clauseCode, alias, receiver, columns)
	}
	return clauseCode
}

func injectColumnReferences(code, alias, receiver string, columns []objects.Column) string {
	if len(columns) == 0 {
		return code
	}
	result := code
	for _, col := range columns {
		quoted := fmt.Sprintf("\"%s\"", col.Name)
		if !strings.Contains(result, quoted) {
			continue
		}
		fieldName := utils.SnakeCaseToPascalCase(col.Name)
		replacement := fmt.Sprintf("%s.ColOf(%s, %s.%s)", alias, receiver, receiver, fieldName)
		result = strings.ReplaceAll(result, quoted, replacement)
	}
	return result
}

func normalizeAliasBooleanClause(code, alias string) string {
	trueLiteral := fmt.Sprintf("%s.Clause(\"TRUE\")", alias)
	if code == trueLiteral {
		return fmt.Sprintf("%s.True", alias)
	}
	falseLiteral := fmt.Sprintf("%s.Clause(\"FALSE\")", alias)
	if code == falseLiteral {
		return fmt.Sprintf("%s.False", alias)
	}
	return code
}

func buildColumnTag(c objects.Column, mapPk map[string]bool, userDefinedType *objects.Type, validationTags state.ModelValidationTag) string {
	var tags []string

	// append json tag
	jsonTag := fmt.Sprintf("json:%q", utils.ToSnakeCase(c.Name)+",omitempty")
	tags = append(tags, jsonTag)

	// append validate tag
	if validationTags != nil {
		if vTag, exist := validationTags[c.Name]; exist {
			tags = append(tags, fmt.Sprintf("validate:%q", vTag))
		}
	}

	// append column tag
	columnTags := []string{
		fmt.Sprintf("name:%s", c.Name),
	}

	if userDefinedType != nil {
		columnTags = append(columnTags, "type:"+string(userDefinedType.Name))
	} else {
		if postgres.IsValidDataType(c.DataType) {
			pdType := postgres.GetPgDataTypeName(postgres.DataType(c.DataType), true)
			columnTags = append(columnTags, "type:"+string(pdType))
		}
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
			columnTags = append(columnTags, "default:"+utils.CleanDoubleColonPattern(defaultStr))
		}
	}

	if c.IsUnique {
		columnTags = append(columnTags, "unique")
	}

	tags = append(tags, fmt.Sprintf("column:%q", strings.Join(columnTags, ";")))

	return strings.Join(tags, " ")
}

func containsRelation(relations []state.Relation, r state.Relation) bool {
	for _, rel := range relations {
		if rel.Tag == r.Tag {
			return true
		}
	}

	return false
}

func BuildRelationTag(r *state.Relation) string {
	var tags []string
	var joinTags []string

	// append json tag
	jsonTag := fmt.Sprintf("json:%q", utils.ToSnakeCase(r.Table)+",omitempty")
	tags = append(tags, jsonTag)

	if r.Action != nil {

		onUpdate, onDelete := objects.RelationActionDefaultLabel, objects.RelationActionDefaultLabel
		if r.Action.UpdateAction != "" {
			code := strings.ToLower(r.Action.UpdateAction)
			if v, ok := objects.RelationActionMapLabel[objects.RelationAction(code)]; ok {
				onUpdate = v
			}
		}

		if r.Action.DeletionAction != "" {
			code := strings.ToLower(r.Action.DeletionAction)
			if v, ok := objects.RelationActionMapLabel[objects.RelationAction(code)]; ok {
				onDelete = v
			}
		}

		tags = append(tags, fmt.Sprintf("onUpdate:%q", onUpdate))
		tags = append(tags, fmt.Sprintf("onDelete:%q", onDelete))
	}

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

func BuildRelationFields(table objects.Table, relations []state.Relation) (mappedRelations []state.Relation) {
	for i := range relations {
		r := relations[i]
		ModelLogger.Trace("generate model relation", "identifier", fmt.Sprintf("%s_%s_%s", r.Table, r.PrimaryKey, r.ForeignKey))
		if r.RelationType == raiden.RelationTypeManyToMany && r.JoinRelation != nil {
			ModelLogger.Trace("generate model relation", "join", fmt.Sprintf("%s_%s_%s_%s", r.Through, r.SourcePrimaryKey, r.TargetPrimaryKey, r.JoinsSourceForeignKey))
		}
		ModelLogger.Trace("generate model relation", "related", r.Type)

		if r.RelationType == raiden.RelationTypeHasOne {
			snakeFk := utils.ToSnakeCase(r.ForeignKey)
			fkTableSplit := strings.Split(snakeFk, "_")
			fkName := inflection.Singular(utils.SnakeCaseToPascalCase(fkTableSplit[0]))

			r.Table = inflection.Singular(utils.SnakeCaseToPascalCase(r.Table))
			if fkName != r.Table {
				r.Table = fmt.Sprintf("%s%s", r.Table, fkName)
			}
		}

		if r.RelationType == raiden.RelationTypeHasMany {
			snakeFk := utils.ToSnakeCase(r.ForeignKey)
			fkTableSplit := strings.Split(snakeFk, "_")
			fkName := inflection.Plural(utils.SnakeCaseToPascalCase(fkTableSplit[0]))
			r.Table = inflection.Plural(utils.SnakeCaseToPascalCase(r.Table))
			if fkName != r.Table {
				r.Table = fmt.Sprintf("%s%s", inflection.Singular(r.Table), fkName)
			}
		}

		if r.JoinRelation != nil {
			throughSuffix := fmt.Sprintf("%s_%s", r.Through, strings.Replace(r.JoinsSourceForeignKey, "_id", "", -1))
			r.Table = fmt.Sprintf("%sThrough%s", inflection.Plural(r.Table), utils.SnakeCaseToPascalCase(inflection.Singular(throughSuffix)))
		}

		r.Tag = BuildRelationTag(&r)

		if !containsRelation(mappedRelations, r) {
			mappedRelations = append(mappedRelations, r)
		}
	}

	sort.Slice(mappedRelations, func(i, j int) bool {
		iToken := mappedRelations[i].Table + mappedRelations[i].Tag + mappedRelations[i].Type
		jToken := mappedRelations[j].Table + mappedRelations[j].Tag + mappedRelations[j].Type
		iRunes := []rune(iToken)
		jRunes := []rune(jToken)

		max := len(iRunes)
		if max > len(jRunes) {
			max = len(jRunes)
		}

		for idx := 0; idx < max; idx++ {
			ir := iRunes[idx]
			jr := jRunes[idx]

			lir := unicode.ToLower(ir)
			ljr := unicode.ToLower(jr)

			if lir != ljr {
				return lir < ljr
			}

			if ir != jr {
				return ir < jr
			}
		}

		return len(iRunes) < len(jRunes)
	})

	return
}
