package generator

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/hashicorp/go-hclog"
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/postgres"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/utils"
)

var RpcLogger hclog.Logger = logger.HcLog().Named("generator.rpc")

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
		Rpc                raiden.RpcBase
		MapScannedTable    map[string]*RpcScannedTable
		OriginalReturnType string
		UseParamPrefix     bool
	}

	GenerateRpcData struct {
		Package string
		Imports []string

		UseParamPrefix bool
		Params         []RpcColumn

		ReturnType   string
		ReturnColumn []RpcColumn
		ReturnDecl   string
		IsReturnArr  bool

		Name     string
		Schema   string
		Security string
		Behavior string

		Models     string
		Definition string
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
	{{ .Field }} {{ .Type }} ` + "`{{ .Tag }}`" + `
	{{- end }}
}

{{- if gt (len .ReturnColumn) 0 }}
type {{ .Name }}Item struct {
	{{- range .ReturnColumn }}
	{{ .Field }} {{ .Type }} ` + "`{{ .Tag }}`" + `
	{{- end }}
}

type {{ .Name }}Result {{ if .IsReturnArr }}[]{{ end }}{{ .Name }}Item
{{- else }}
type {{ .Name }}Result {{ if .IsReturnArr }}[]{{ end }}{{ .ReturnDecl }}
{{- end }}

type {{ .Name }} struct {
	raiden.RpcBase
	Params   *{{ .Name }}Params ` + "`json:\"-\"`" + `
	Return   {{ .Name }}Result ` + "`json:\"-\"`" + `
}

func (r *{{ .Name }}) GetName() string {
	return "{{ .Name | ToSnakeCase }}"
}

{{- if ne .Schema "public" }}

func (r *{{.Name }}) GetSchema() string {
	return "{{ .Schema }}"
}

{{- end }}
{{- if not .UseParamPrefix }}

func  (r *{{.Name }}) UseParamPrefix() bool {
	return false
}

{{- end }}
{{- if ne .Security "RpcSecurityTypeInvoker" }}

func (r *{{.Name }}) GetSecurity() raiden.RpcSecurityType {
	return raiden.{{ .Security }}
}
{{- end }}
{{- if ne .Behavior "RpcBehaviorVolatile" }}

func (r *{{.Name }}) GetBehavior() raiden.RpcBehaviorType {
	return raiden.{{ .Behavior }}
}
{{- end }}

func (r *{{.Name }}) GetReturnType() raiden.RpcReturnDataType {
	return raiden.{{ .ReturnType }}
}
{{- if ne .Models "" }}

func (r *{{ .Name }}) BindModels() {
	{{ .Models }}
}
{{- end }}

func (r *{{ .Name }}) GetRawDefinition() string {
	return ` + "`{{ .Definition }}`" + `
}`
)

func GenerateRpc(basePath string, projectName string, functions []objects.Function, generateFn GenerateFn) (err error) {
	folderPath := filepath.Join(basePath, RpcDir)
	RpcLogger.Trace("create rpc folder if not exist", "path", folderPath)
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

func generateRpcItem(folderPath string, projectName string, function *objects.Function, generateFn GenerateFn) error {
	// define binding func
	funcMaps := []template.FuncMap{
		{"ToSnakeCase": utils.ToSnakeCase},
	}

	// set imports path

	raidenPath := fmt.Sprintf("%q", "github.com/sev-2/raiden")
	importsMap := map[string]bool{
		raidenPath: true,
	}

	// define file path
	filePath := filepath.Join(folderPath, fmt.Sprintf("%s.%s", utils.ToSnakeCase(function.Name), "go"))

	// // extract rpc function
	result, err := ExtractRpcFunction(function)
	if err != nil {
		return err
	}

	rpcParams, err := result.GetParams(importsMap)
	if err != nil {
		return err
	}

	returnDecl, returnColumns, IsReturnArr, err := result.GetReturn(importsMap)
	if err != nil {
		return err
	}

	returnTypeDecl, err := raiden.GetValidRpcReturnNameDecl(result.Rpc.ReturnType, true)
	if err != nil {
		return err
	}

	if result.GetModelDecl() != "" {
		modelsImportPath := fmt.Sprintf("%s/%s", utils.ToGoModuleName(projectName), ModelDir)
		modePath := fmt.Sprintf("%q", modelsImportPath)
		importsMap[modePath] = true
	}

	var importsPath []string
	for key := range importsMap {
		importsPath = append(importsPath, key)
	}

	// set data
	data := GenerateRpcData{
		Package:        "rpc",
		Imports:        importsPath,
		Name:           utils.SnakeCaseToPascalCase(function.Name),
		Params:         rpcParams,
		UseParamPrefix: result.UseParamPrefix,
		ReturnType:     returnTypeDecl,
		ReturnDecl:     returnDecl,
		ReturnColumn:   returnColumns,
		IsReturnArr:    IsReturnArr,
		Schema:         result.Rpc.Schema,
		Security:       result.GetSecurity(),
		Behavior:       result.GetBehavior(),
		Models:         result.GetModelDecl(),
		Definition:     result.Rpc.Definition,
	}

	// setup generate input param
	generateInput := GenerateInput{
		BindData:     data,
		FuncMap:      funcMaps,
		Template:     RpcTemplate,
		TemplateName: "rpcTemplate",
		OutputPath:   filePath,
	}

	RpcLogger.Debug("generate rpc", "path", generateInput.OutputPath)
	return generateFn(generateInput, nil)
}

func ExtractRpcFunction(fn *objects.Function) (result ExtractRpcDataResult, err error) {
	//  extract param
	params, usePrefix, e := ExtractRpcParam(fn)
	if e != nil {
		err = e
		return
	}

	// get model and definition
	cleanDef := strings.ReplaceAll(fn.Definition, "\\n", "")
	definition, mapScannedTable, e := ExtractRpcTable(cleanDef)
	if e != nil {
		err = e
		return
	}

	// normalize aliases
	if e := RpcNormalizeTableAliases(mapScannedTable); e != nil {
		err = e
		return
	}

	// set return type
	securityType := raiden.RpcSecurityTypeInvoker
	if fn.SecurityDefiner {
		securityType = raiden.RpcSecurityTypeDefiner
	}

	returnType := raiden.RpcReturnDataTypeVoid
	returnTypeLc := strings.ToLower(fn.ReturnType)
	if strings.Contains(returnTypeLc, "setof") {
		returnType = raiden.RpcReturnDataTypeSetOf
	} else if strings.Contains(returnTypeLc, "table") {
		returnType = raiden.RpcReturnDataTypeTable
	} else {
		returnType, err = raiden.GetValidRpcReturnType(fn.ReturnType, true)
		if err != nil {
			return
		}
	}

	// assign rpc params
	result.Rpc.Params = params
	result.Rpc.Schema = fn.Schema
	result.Rpc.Behavior = raiden.RpcBehaviorType(fn.Behavior)
	result.Rpc.Name = fn.Name
	result.Rpc.Definition = bindModelToDefinition(definition, mapScannedTable, result.Rpc.Params, usePrefix)
	result.Rpc.CompleteStatement = fn.CompleteStatement
	result.Rpc.SecurityType = securityType
	result.Rpc.ReturnType = returnType

	result.OriginalReturnType = fn.ReturnType
	result.MapScannedTable = mapScannedTable
	result.UseParamPrefix = usePrefix

	return
}

func ExtractRpcParam(fn *objects.Function) (params []raiden.RpcParam, usePrefix bool, err error) {
	mapParam := make(map[string]string)

	// bind param to map
	// example argument value :
	// - "in_candidate_name character varying DEFAULT 'anon'::character varying, in_voter_name character varying DEFAULT 'anon'::character varying"
	for _, at := range strings.Split(fn.ArgumentTypes, ",") {
		// at example : in_candidate_name character varying DEFAULT 'anon'::character varying
		cleanArgType := strings.TrimLeft(strings.TrimRight(at, " "), " ")
		if strings.Contains(cleanArgType, "DEFAULT") {
			splitTextDefault := strings.Split(cleanArgType, "DEFAULT")
			if len(splitTextDefault) != 2 {
				continue
			}

			// update param text type
			// splitTextDefault[0] example : in_candidate_name character varying
			typeParamStr := strings.TrimLeft(strings.TrimRight(splitTextDefault[0], " "), " ")

			// example splitIat result :  ["in_candidate_name", "character varying"]
			splitIat := strings.SplitN(typeParamStr, " ", 2)
			if len(splitIat) >= 2 {
				// example : mapParam["in_candidate_name"] =  "character varying"
				mapParam[splitIat[0]] = splitIat[1]
			}

			// bind default value to map
			// splitTextDefault[1] example : 'anon'::character varyings
			defaultTextArr := strings.Split(splitTextDefault[1], "::")
			if len(defaultTextArr) > 0 {
				// example : mapParam["in_candidate_name_default"] =  "anon"
				mapParam[splitIat[0]+"_default"] = strings.ReplaceAll(
					strings.ReplaceAll(
						strings.TrimLeft(
							strings.TrimRight(defaultTextArr[0], " "), " "),
						`"`, ""),
					"'", "")
			}
			continue
		}

		splitIat := strings.SplitN(cleanArgType, " ", 2)
		if len(splitIat) == 2 {
			mapParam[splitIat[0]] = splitIat[1]
		}
	}

	// loop for create rpc param and add to params variable
	paramsUsePrefix := []string{}
	paramsInCount := 0
	for i := range fn.Args {
		fa := fn.Args[i]
		if fa.Mode != "in" {
			continue
		}

		paramsInCount++
		if strings.HasPrefix(strings.ToLower(fa.Name), raiden.DefaultRpcParamPrefix) {
			paramsUsePrefix = append(paramsUsePrefix, fa.Name)
		}

		fieldName := strings.ReplaceAll(fa.Name, raiden.DefaultRpcParamPrefix, "")
		p := raiden.RpcParam{
			Name: fieldName,
			Type: raiden.RpcParamDataTypeText,
		}

		// get data type
		if pt, isParamExist := mapParam[fa.Name]; isParamExist {
			pt = strings.TrimLeft(strings.TrimRight(pt, " "), " ")
			paramType, err := raiden.GetValidRpcParamType(pt, true)
			if err != nil {
				return params, usePrefix, err
			}
			p.Type = paramType
		}

		// get default value
		if d, isDefaultExist := mapParam[fa.Name+"_default"]; isDefaultExist && fa.HasDefault {
			p.Default = &d
		}
		params = append(params, p)
	}

	usePrefix = len(paramsUsePrefix) == paramsInCount

	return
}

// code for detect table in query and return array of object contain table name, table aliases and relation
func ExtractRpcTable(def string) (string, map[string]*RpcScannedTable, error) {
	dFields := strings.Fields(utils.CleanUpString(def))
	mapResult := make(map[string]*RpcScannedTable)
	mapTableOrAlias := make(map[string]string)

	// extract table name
	var lastField string
	var foundTable = &RpcScannedTable{}

	// value true if command start with create, update, delete, alter, drop, alter, truncate and etc
	// value false if command start with select or with
	var writeMode = false

	for _, f := range dFields {
		f = strings.TrimRight(f, ";")
		k := strings.ToUpper(f)

		if k == postgres.Select || k == postgres.With {
			writeMode = false
		} else {
			writeMode = true
		}

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
				if postgres.IsReservedSymbol(f) {
					continue
				}
				foundTable.Alias = f
			}
		case postgres.Inner, postgres.Outer, postgres.Left, postgres.Right:
			if k == postgres.Join {
				lastField += " " + postgres.Join
				continue
			}
		case postgres.Join, postgres.InnerJoin, postgres.OuterJoin, postgres.LeftJoin, postgres.RightJoin:
			if k == postgres.On {
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
				if postgres.IsReservedSymbol(f) {
					continue
				}
				foundTable.Alias = f
			}
		case postgres.On:
			if writeMode {
				lastField = k
				continue
			}

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

func bindModelToDefinition(def string, mapTable map[string]*RpcScannedTable, params []raiden.RpcParam, useParamPrefix bool) string {
	for _, v := range mapTable {
		pattern := fmt.Sprintf(`\b%s\b`, regexp.QuoteMeta(v.Name))
		def = regexp.MustCompile(pattern).ReplaceAllString(def, ":"+v.Alias)
	}

	for i := range params {
		p := params[i]
		findKey, replaceKey := p.Name, p.Name
		if useParamPrefix {
			findKey = raiden.DefaultRpcParamPrefix + findKey
		}
		pattern := fmt.Sprintf(`\b%s\b`, regexp.QuoteMeta(findKey))
		def = regexp.MustCompile(pattern).ReplaceAllString(def, ":"+replaceKey)
	}
	return def
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

func (r *ExtractRpcDataResult) GetParams(mapImports map[string]bool) (columns []RpcColumn, err error) {
	for i := range r.Rpc.Params {
		p := r.Rpc.Params[i]

		tag := raiden.RpcParamTag{
			Name: p.Name,
			Type: string(p.Type),
		}

		if p.Default != nil {
			tag.DefaultValue = *p.Default
		}

		rpcTag, e := raiden.MarshalRpcParamTag(&tag)
		if e != nil {
			err = e
			return
		}

		c := RpcColumn{
			Field: utils.SnakeCaseToPascalCase(p.Name),
			Type:  raiden.RpcParamToGoType(p.Type),
			Tag:   fmt.Sprintf("json:%q column:%q", p.Name, rpcTag),
		}

		splitType := strings.Split(c.Type, ".")
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
			mapImports[key] = true
		}

		columns = append(columns, c)
	}
	return
}

func (r *ExtractRpcDataResult) GetModelDecl() (modelDecl string) {
	if len(r.MapScannedTable) == 0 {
		return
	}

	var bindModelDeclArr []string
	for _, v := range r.MapScannedTable {
		bindModelDeclArr = append(bindModelDeclArr, fmt.Sprintf("BindModel(models.%s{}, %q)", utils.SnakeCaseToPascalCase(v.Name), v.Alias))
	}
	return "r." + strings.Join(bindModelDeclArr, ".")
}

func (r *ExtractRpcDataResult) GetReturn(mapImports map[string]bool) (returnDecl string, returnColumns []RpcColumn, isReturnArr bool, err error) {
	// set result decl
	frCheck := strings.ToLower(r.OriginalReturnType)
	switch r.Rpc.ReturnType {
	case raiden.RpcReturnDataTypeSetOf:

		split := strings.SplitN(frCheck, " ", 2)
		if len(split) != 2 {
			if err = fmt.Errorf("invalid return type for rpc : %s", r.Rpc.Name); err != nil {
				return
			}
		}

		tableName := split[1]
		_, isExist := r.MapScannedTable[tableName]
		if !isExist {
			err = fmt.Errorf("table %s is not declare in definition function of rpc %s", tableName, r.Rpc.Name)
			return
		}

		isReturnArr = true
		returnDecl = fmt.Sprintf("models.%s", utils.SnakeCaseToPascalCase(tableName))
	case raiden.RpcReturnDataTypeTable:
		// example : "TABLE(id integer, created_at timestamp without time zone, sc_name character varying, c_name character varying)"
		rsType := strings.TrimRight(strings.ReplaceAll(frCheck, "table(", ""), ")")
		rsSplit := strings.Split(rsType, ",")
		for _, v := range rsSplit {
			splitC := strings.SplitN(strings.TrimLeft(v, " "), " ", 2)
			if len(splitC) != 2 {
				continue
			}
			cName := splitC[0]
			cType, e := raiden.GetValidRpcParamType(splitC[1], true)
			if e != nil {
				err = e
				return
			}

			tag := raiden.RpcParamTag{
				Name: cName,
				Type: string(cType),
			}

			rpcTag, e := raiden.MarshalRpcParamTag(&tag)
			if e != nil {
				err = e
				return
			}

			c := RpcColumn{
				Field: utils.SnakeCaseToPascalCase(cName),
				Type:  raiden.RpcParamToGoType(cType),
				Tag:   fmt.Sprintf("json:%q column:%q", cName, rpcTag),
			}

			splitType := strings.Split(c.Type, ".")
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
				mapImports[key] = true
			}

			isReturnArr = true
			returnColumns = append(returnColumns, c)
		}

	default:
		isReturnArr = false
		returnName := strings.ToUpper(string(r.OriginalReturnType))
		returnDecl = raiden.RpcReturnToGoType(raiden.RpcReturnDataType(returnName))
	}

	return
}

func (r *ExtractRpcDataResult) GetSecurity() (security string) {
	switch r.Rpc.SecurityType {
	case raiden.RpcSecurityTypeDefiner:
		return "RpcSecurityTypeDefiner"
	default:
		return "RpcSecurityTypeInvoker"
	}
}

func (r *ExtractRpcDataResult) GetBehavior() (behavior string) {
	switch r.Rpc.Behavior {
	case raiden.RpcBehaviorImmutable:
		return "RpcBehaviorImmutable"
	case raiden.RpcBehaviorStable:
		return "RpcBehaviorStable"
	default:
		return "RpcBehaviorVolatile"
	}
}
