package generator

import (
	"fmt"
	"reflect"
	"strings"
	"text/template"

	"github.com/sev-2/raiden/pkg/postgres"
	"github.com/sev-2/raiden/pkg/supabase"
	"github.com/sev-2/raiden/pkg/utils"
)

// ----- Define type, variable and constant -----
type GenerateInput struct {
	BindData     any
	Template     string
	TemplateName string
	OutputPath   string
	FuncMap      []template.FuncMap
}

type GenerateFn func(input GenerateInput) error

// ----- Generate functionality  -----
func Generate(input GenerateInput) error {
	file, err := utils.CreateFile(input.OutputPath, true)
	if err != nil {
		return fmt.Errorf("failed create file %s : %v", input.OutputPath, err)
	}
	defer file.Close()

	tmplInstance := template.New(input.TemplateName)
	for _, tm := range input.FuncMap {
		tmplInstance.Funcs(tm)
	}

	tmpl, err := tmplInstance.Parse(input.Template)
	if err != nil {
		return fmt.Errorf("error parsing : %v", err)
	}

	err = tmpl.Execute(file, input.BindData)
	if err != nil {
		return fmt.Errorf("error executing : %v", err)
	}

	return nil
}

func generateArrayDeclaration(value reflect.Value, withoutQuote bool) string {
	var arrayValues []string
	for i := 0; i < value.Len(); i++ {
		if withoutQuote {
			arrayValues = append(arrayValues, fmt.Sprintf("%s", value.Index(i).Interface()))
		} else {
			arrayValues = append(arrayValues, fmt.Sprintf("%q", value.Index(i).Interface()))
		}
	}
	return "[]string{" + strings.Join(arrayValues, ", ") + "}"
}

func generateMapDeclarationFromValue(value reflect.Value) string {
	// Start the map declaration
	var resultArr []string
	for _, key := range value.MapKeys() {
		valueStr := valueToString(value.Interface())
		resultArr = append(resultArr, fmt.Sprintf(`%q: %s,`, key, valueStr))
	}

	return "map[string]any{" + strings.Join(resultArr, ", ") + "}"
}

func generateMapDeclaration(mapData map[string]any) string {
	// Start the map declaration
	var resultArr []string
	for key, value := range mapData {
		valueStr := valueToString(value)
		resultArr = append(resultArr, fmt.Sprintf(`%q: %s`, key, valueStr))
	}

	return "map[string]any{" + strings.Join(resultArr, ",") + "}"
}

func generateStructDataType(value reflect.Value) string {
	return fmt.Sprintf("%v", value.Interface())
}

func valueToString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return fmt.Sprintf("%q", v)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%v", v)
	case float32, float64:
		return fmt.Sprintf("%v", v)
	case bool:
		return fmt.Sprintf("%t", v)
	default:
		return fmt.Sprintf(`"%v"`, v) // Default: use fmt.Sprint for other types
	}
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
