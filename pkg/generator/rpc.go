package generator

import (
	"fmt"
	"path/filepath"
	"regexp"
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
	RpcScannedTable struct {
		Name     string
		Alias    string
		Relation []string
	}

	RpcColumn struct {
		Field string
		Type  string
		Tag   string
	}

	ExtractRpcDataResult struct {
		Definition       string
		BindModelDecl    string
		ResultDecl       string
		ResultTag        string
		ResultColumn     []RpcColumn
		IsResultArr      bool
		ImportResultPath []string
		MetadataTag      string
	}

	GenerateRpcData struct {
		Definition   string
		Imports      []string
		Package      string
		Params       []RpcColumn
		MetadataTag  string
		Models       string
		Name         string
		ResultDecl   string
		ResultTag    string
		ResultColumn []RpcColumn
		IsResultArr  bool
	}
)

const (
	RpcDir      = "internal/rpc"
	RpcTemplate = `package {{ .Package }}
{{- if gt (len .Imports) 0 }}

import (
{{- range .Imports}}
	{{.}}
{{- end}}
)
{{- end }}

type {{ .Name }}Params struct {
	{{- range .Params }}
	{{ .Field | ToGoIdentifier }} {{ .Type }} ` + "`{{ .Tag }}`" + `
	{{- end }}
}

{{- if gt (len .ResultColumn) 0 }}
type {{ .Name }}Item struct {
	{{- range .ResultColumn }}
	{{ .Field | ToGoIdentifier }} {{ .Type }} ` + "`{{ .Tag }}`" + `
	{{- end }}
}

type {{ .Name }}Result {{ if .IsResultArr }}[]{{ end }}{{ .Name }}Item
{{- else }}
type {{ .Name }}Result {{ if .IsResultArr }}[]{{ end }}{{ .ResultDecl }}
{{- end }}

type {{ .Name }} struct {
	raiden.RpcBase
	Params   {{ .Name }}Params
	Result   {{ .Name }}Result ` + "`json:\"-\" {{ .ResultTag }}`" + `
	Metadata string ` + "`json:\"-\" {{ .MetadataTag }}`" + `
}

func (r *{{ .Name }}) BindModels() {
	{{ .Models }}
}

func (r *{{ .Name }}) GetDefinition() string {
	return ` + "`{{ .Definition }}`" + `
}`
)

func GenerateRpc(basePath string, projectName string, functions []supabase.Function, generateFn GenerateFn) (err error) {
	folderPath := filepath.Join(basePath, RpcDir)
	logger.Debugf("GenerateRpc - create %s folder if not exist", folderPath)
	if exist := utils.IsFolderExists(folderPath); !exist {
		if err := utils.CreateFolder(folderPath); err != nil {
			return err
		}
	}

	for i := range functions {
		f := functions[i]
		if err := generateRpcItem(folderPath, projectName, &f, generateFn); err != nil {
			return err
		}
	}

	return nil
}

func generateRpcItem(folderPath string, projectName string, function *supabase.Function, generateFn GenerateFn) error {
	// define binding func
	funcMaps := []template.FuncMap{
		{"ToGoIdentifier": utils.SnakeCaseToPascalCase},
	}

	// set imports path
	modelsImportPath := fmt.Sprintf("%s/%s", utils.ToGoModuleName(projectName), ModelDir)
	modePath := fmt.Sprintf("%q", modelsImportPath)
	raidenPath := fmt.Sprintf("%q", "github.com/sev-2/raiden")
	importsMap := map[string]bool{
		raidenPath: true,
		modePath:   true,
	}

	// define file path
	filePath := filepath.Join(folderPath, fmt.Sprintf("%s.%s", utils.ToSnakeCase(function.Name), "go"))

	// extract rpc params
	params, err := ExtractRpcParam(function, importsMap)
	if err != nil {
		return err
	}

	// extract rpc function
	resultExtract, err := ExtractRpcData(function, importsMap)
	if err != nil {
		return err
	}

	var importsPath []string
	for key := range importsMap {
		importsPath = append(importsPath, key)
	}

	// set data
	data := GenerateRpcData{
		Package:      "rpc",
		Imports:      importsPath,
		Name:         utils.SnakeCaseToPascalCase(function.Name),
		Definition:   resultExtract.Definition,
		Params:       params,
		MetadataTag:  resultExtract.MetadataTag,
		Models:       resultExtract.BindModelDecl,
		ResultDecl:   resultExtract.ResultDecl,
		ResultTag:    resultExtract.ResultTag,
		ResultColumn: resultExtract.ResultColumn,
		IsResultArr:  resultExtract.IsResultArr,
	}

	// setup generate input param
	generateInput := GenerateInput{
		BindData:     data,
		FuncMap:      funcMaps,
		Template:     RpcTemplate,
		TemplateName: "rpcTemplate",
		OutputPath:   filePath,
	}

	logger.Debugf("GenerateRpc - generate rpc to %s", generateInput.OutputPath)
	return generateFn(generateInput, nil)
}

func ExtractRpcParam(fn *supabase.Function, importsMap map[string]bool) (params []RpcColumn, err error) {
	mapParam := make(map[string]string)

	// bind param to map
	// example argument value :
	// - "in_candidate_name character varying DEFAULT 'anon'::character varying, in_voter_name character varying DEFAULT 'anon'::character varying"
	for _, at := range strings.Split(fn.ArgumentTypes, ",") {
		if strings.Contains(at, "DEFAULT") {
			splitTextDefault := strings.Split(at, "DEFAULT")
			if len(splitTextDefault) != 2 {
				continue
			}

			// update param text type
			typeParamStr := strings.TrimLeft(strings.TrimRight(splitTextDefault[0], " "), " ")
			splitIat := strings.SplitN(typeParamStr, " ", 2)
			if len(splitIat) >= 2 {
				mapParam[splitIat[0]] = splitIat[1]
			}

			// bind default value to map
			defaultTextArr := strings.Split(splitTextDefault[1], "::")
			if len(defaultTextArr) > 0 {
				mapParam[splitIat[0]+"_default"] = strings.ReplaceAll(
					strings.ReplaceAll(
						strings.TrimLeft(
							strings.TrimRight(defaultTextArr[0], " "), " "),
						`"`, ""),
					"'", "")
			}
			continue
		}

		splitIat := strings.SplitN(at, " ", 2)
		if len(splitIat) == 2 {
			mapParam[splitIat[0]] = splitIat[1]
		}

	}

	// loop for create rpc param and add to params variable
	for i := range fn.Args {
		fa := fn.Args[i]
		if fa.Mode != "in" {
			continue
		}

		fieldName := strings.ReplaceAll(fa.Name, raiden.DefaultParamPrefix, "")
		p := RpcColumn{
			Field: utils.SnakeCaseToPascalCase(fieldName),
			Type:  "string",
		}
		pTag := raiden.RpcParamTag{
			Name: fieldName,
		}

		// get data type
		if pt, isParamExist := mapParam[fa.Name]; isParamExist {
			pt = strings.TrimLeft(strings.TrimRight(pt, " "), " ")
			paramType, err := raiden.GetValidRpcParamType(pt, true)
			if err != nil {
				return params, err
			}
			p.Type = raiden.RpcParamToGoType(paramType)
			pTag.Type = string(paramType)

			// update import path
			splitType := strings.Split(p.Type, ".")
			if len(splitType) > 1 {
				importPackage := splitType[0]

				var importPackageName string
				switch importPackage {
				case "time":
					importPackageName = importPackage
				case "uuid":
					importPackageName = "github.com/google/uuid"
				case "json":
					importPackageName = "encoding/json"
				}
				key := fmt.Sprintf("%q", importPackageName)
				importsMap[key] = true
			}
		}

		// get default value
		if d, isDefaultExist := mapParam[fa.Name+"_default"]; isDefaultExist && fa.HasDefault {
			pTag.DefaultValue = d
		}

		// create param tag
		ptStr, err := raiden.MarshalRpcParamTag(&pTag)
		if err != nil {
			return params, err
		}
		tagArr := []string{
			fmt.Sprintf("json:%q", fieldName),
			fmt.Sprintf("param:%q", ptStr),
		}
		p.Tag = strings.Join(tagArr, " ")

		params = append(params, p)
	}

	return
}

func ExtractRpcData(fn *supabase.Function, importsMap map[string]bool) (result ExtractRpcDataResult, err error) {
	// set default definition
	result.Definition = strings.ReplaceAll(fn.Definition, "\\n", "")

	// extract table
	def, mapTable, e := ExtractRpcTable(result.Definition)
	if e != nil {
		err = e
		return
	}

	// make sure all table have binding alias
	if e := RpcNormalizeTableAliases(mapTable); e != nil {
		err = e
		return
	}

	if err = bindRpcFunctionResult(fn, mapTable, &result, importsMap); err != nil {
		return
	}

	bindModelDecl(def, mapTable, &result)

	// update definitions and set bind model decl

	err = bindRpcMetadataTag(fn, &result)
	return
}

func ExtractRpcTable(def string) (string, map[string]*RpcScannedTable, error) {
	dFields := strings.Fields(utils.CleanUpString(def))
	mapResult := make(map[string]*RpcScannedTable)
	mapTableOrAlias := make(map[string]string)
	// extract table name
	var lastField string
	var foundTable = &RpcScannedTable{}
	for _, f := range dFields {
		k := strings.ToUpper(f)
		switch lastField {
		case postgres.From:
			if postgres.IsReservedKeyword(k) {
				mapResult[foundTable.Name] = foundTable
				mapTableOrAlias[foundTable.Name] = foundTable.Name
				if foundTable.Alias != "" {
					mapTableOrAlias[foundTable.Alias] = foundTable.Name
				}
				foundTable = &RpcScannedTable{}
				lastField = k
				continue
			}
			if len(foundTable.Name) == 0 {
				split := strings.Split(f, ".")
				if len(split) == 2 {
					foundTable.Name = split[1]
					continue
				}
				foundTable.Name = f
			} else {
				foundTable.Alias = f
			}
		case postgres.Join:
			if f == postgres.On {
				lastField = f
				continue
			}

			if len(foundTable.Name) == 0 {
				split := strings.Split(f, ".")
				if len(split) == 2 {
					foundTable.Name = split[1]
					continue
				}
				foundTable.Name = f
			} else {
				foundTable.Alias = f
			}
		case postgres.On:
			if postgres.IsReservedKeyword(k) {
				mapResult[foundTable.Name] = foundTable
				mapTableOrAlias[foundTable.Name] = foundTable.Name
				if foundTable.Alias != "" {
					mapTableOrAlias[foundTable.Alias] = foundTable.Name
				}
				foundTable = &RpcScannedTable{}
				lastField = k
				continue
			}

			if f == "=" {
				continue
			}

			splitRelationKey := strings.Split(f, ".")
			if splitRelationKey[0] == foundTable.Name || splitRelationKey[0] == foundTable.Alias {
				continue
			}

			if len(splitRelationKey) == 3 {
				splitRelationKey[0] = splitRelationKey[1]
				splitRelationKey[1] = splitRelationKey[2]
			}

			relationTable, isTableExist := mapTableOrAlias[splitRelationKey[0]]
			if !isTableExist || relationTable == "" {
				rt, isAliasExist := mapTableOrAlias[splitRelationKey[0]]
				if !isAliasExist || rt == "" {
					return "", nil, fmt.Errorf("table %s is not exist", splitRelationKey[0])
				}
				relationTable = rt
			}

			foundTable.Relation = append(foundTable.Relation, relationTable)

			foundRelationTable := mapResult[relationTable]
			if foundRelationTable.Name != "" {
				var isRelationExist bool
				for _, v := range foundRelationTable.Relation {
					if v == foundTable.Name {
						isRelationExist = true
						break
					}
				}

				if !isRelationExist {
					foundRelationTable.Relation = append(foundRelationTable.Relation, foundTable.Name)
				}
			}

		default:
			lastField = k
		}
	}

	return strings.Join(dFields, " "), mapResult, nil
}

func RpcNormalizeTableAliases(mapTables map[string]*RpcScannedTable) error {
	mapAlias := make(map[string]bool)
	for _, v := range mapTables {
		if v.Alias != "" && v.Name != "" {
			mapAlias[v.Alias] = true
		}
	}

	for _, v := range mapTables {
		if v.Alias != "" && v.Name != "" {
			continue
		}
		newAlias := findAvailableAlias(v.Name, mapAlias, 1)
		if newAlias == "" {
			return fmt.Errorf("cannot generate alias for table : %s", v.Name)
		}
		v.Alias = newAlias
		mapAlias[newAlias] = true
	}

	return nil
}

func findAvailableAlias(tableName string, mapAlias map[string]bool, sub int) (alias string) {
	if len(tableName) == sub {
		return ""
	}

	newAlias := tableName[:sub]
	_, isExist := mapAlias[newAlias]
	if !isExist {
		return newAlias
	}

	return findAvailableAlias(tableName, mapAlias, sub+1)
}

func bindRpcMetadataTag(fn *supabase.Function, result *ExtractRpcDataResult) error {
	// set metadata
	metadataTagInstance := raiden.RpcMetadataTag{
		Name:     fn.Name,
		Schema:   fn.Schema,
		Security: raiden.RpcSecurityTypeInvoker,
		Behavior: raiden.RpcBehaviorVolatile,
	}

	// set metadata security definer
	if fn.SecurityDefiner {
		metadataTagInstance.Security = raiden.RpcSecurityTypeDefiner
	}

	// set metadata behavior
	switch strings.ToUpper(fn.Behavior) {
	case string(raiden.RpcBehaviorVolatile):
		metadataTagInstance.Behavior = raiden.RpcBehaviorVolatile
	case string(raiden.RpcBehaviorStable):
		metadataTagInstance.Behavior = raiden.RpcBehaviorStable
	case string(raiden.RpcBehaviorImmutable):
		metadataTagInstance.Behavior = raiden.RpcBehaviorImmutable
	}

	tag, err := raiden.MarshalRpcMetadataTag(&metadataTagInstance)
	if err != nil {
		return err
	}

	result.MetadataTag = tag
	return nil
}

func bindRpcFunctionResult(fn *supabase.Function, mapTable map[string]*RpcScannedTable, result *ExtractRpcDataResult, importsMap map[string]bool) error {
	// set result decl
	frCheck := strings.ToLower(fn.ReturnType)
	if strings.Contains(frCheck, "setof") {
		// example
		split := strings.SplitN(frCheck, " ", 2)
		if len(split) != 2 {
			if err := fmt.Errorf("invalid return type for rpc : %s", fn.Name); err != nil {
				return err
			}
		}

		tableName := split[1]
		_, isExist := mapTable[tableName]
		if !isExist {
			err := fmt.Errorf("table %s is not declare in definition function of rpc %s", tableName, fn.Name)
			return err
		}

		result.ResultDecl = fmt.Sprintf("models.%s", utils.SnakeCaseToPascalCase(tableName))
		result.ResultTag = fmt.Sprintf("returnType:%q", strings.ToLower(fn.ReturnType))
		result.IsResultArr = true
	} else if strings.Contains(frCheck, "table") {
		// example : "TABLE(id integer, created_at timestamp without time zone, sc_name character varying, c_name character varying)"
		result.IsResultArr = true
		rsType := strings.TrimRight(strings.ReplaceAll(frCheck, "table(", ""), ")")
		rsSplit := strings.Split(rsType, ",")

		resultColumn := make([]RpcColumn, 0)
		importsPath := make([]string, 0)
		for _, v := range rsSplit {
			splitC := strings.SplitN(strings.TrimLeft(v, " "), " ", 2)
			if len(splitC) == 2 {
				cType, err := raiden.GetValidRpcParamType(splitC[1], true)
				if err != nil {
					return err
				}

				rsColum := RpcColumn{
					Field: utils.SnakeCaseToPascalCase(splitC[0]),
					Type:  raiden.RpcReturnToGoType(raiden.RpcReturnDataType(cType)),
					Tag:   fmt.Sprintf("json:%q", utils.ToSnakeCase(splitC[0])),
				}

				splitType := strings.Split(rsColum.Type, ".")
				if len(splitType) > 1 {
					importPackage := splitType[0]
					var importPackageName string
					switch importPackage {
					case "time":
						importPackageName = importPackage
					case "uuid":
						importPackageName = "github.com/google/uuid"
					case "json":
						importPackageName = "encoding/json"
					}
					key := fmt.Sprintf("%q", importPackageName)
					importsMap[key] = true
				}

				resultColumn = append(resultColumn, rsColum)
			}
		}
		result.ImportResultPath = append(result.ImportResultPath, importsPath...)
		result.ResultTag = fmt.Sprintf("returnType:%q", strings.ToLower(fn.ReturnType))
		result.ResultColumn = resultColumn
	} else {
		result.IsResultArr = false
		returnName := strings.ToUpper(fn.ReturnType)
		result.ResultDecl = raiden.RpcReturnToGoType(raiden.RpcReturnDataType(returnName))
		result.ResultTag = fmt.Sprintf("returnType:%q", fn.ReturnType)
	}

	return nil
}

func bindModelDecl(def string, mapTable map[string]*RpcScannedTable, result *ExtractRpcDataResult) {
	var bindModelDeclArr []string

	for _, v := range mapTable {
		bindModelDeclArr = append(bindModelDeclArr, fmt.Sprintf("BindModel(models.%s{}, %q)", utils.SnakeCaseToPascalCase(v.Name), v.Alias))
		pattern := fmt.Sprintf(`\b%s\b`, regexp.QuoteMeta(v.Name))
		def = regexp.MustCompile(pattern).ReplaceAllString(def, ":"+v.Alias)
	}
	result.Definition = def
	result.BindModelDecl = "r." + strings.Join(bindModelDeclArr, ".")
}
