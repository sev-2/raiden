package generator

import (
	"fmt"
	"path/filepath"
	"reflect"
	"text/template"

	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/supabase"
	"github.com/sev-2/raiden/pkg/utils"
)

// ----- Define type, var and const -----
type GenerateRoleData struct {
	Fields    []map[string]any
	Imports   []string
	Package   string
	RoleName  string
	NumFields int
}

const (
	RoleDir      = "internal/roles"
	RoleTemplate = `package {{ .Package }}
{{- if gt (len .Imports) 0 }}

import (
{{- range .Imports}}
	{{.}}
{{- end}}
)
{{- end }}

var {{ .RoleName | ToGoIdentifier }} = &postgres.Role{
{{- range $i, $field := .Fields }}
	{{ .Name }} : {{ .Value }},
{{- end }}
}
`
)

func GenerateRoles(basePath string, roles []supabase.Role, generateFn GenerateFn) (err error) {
	folderPath := filepath.Join(basePath, RoleDir)
	logger.Debugf("GenerateRoles - create %s folder if not exist", folderPath)
	if exist := utils.IsFolderExists(folderPath); !exist {
		if err := utils.CreateFolder(folderPath); err != nil {
			return err
		}
	}

	for _, v := range roles {
		if err := GenerateRole(folderPath, v, generateFn); err != nil {
			return err
		}
	}

	return nil
}

func GenerateRole(folderPath string, role supabase.Role, generateFn GenerateFn) error {
	// define binding func
	funcMaps := []template.FuncMap{
		{"ToGoIdentifier": utils.SnakeCaseToPascalCase},
	}

	// define file path
	filePath := filepath.Join(folderPath, fmt.Sprintf("%s.%s", role.Name, "go"))

	// generate fields
	fields := getRoleFields(role)

	// set imports path
	imports := []string{
		fmt.Sprintf("%q", "github.com/sev-2/raiden/pkg/postgres"),
	}

	// execute the template and write to the file
	data := GenerateRoleData{
		Package:   "roles",
		Imports:   imports,
		RoleName:  role.Name,
		Fields:    fields,
		NumFields: len(fields),
	}

	// set input
	input := GenerateInput{
		BindData:     data,
		Template:     RoleTemplate,
		TemplateName: "roleTemplate",
		OutputPath:   filePath,
		FuncMap:      funcMaps,
	}

	logger.Debugf("GenerateRoles - generate role to %s", input.OutputPath)
	return generateFn(input)
}

func getRoleFields(role supabase.Role) (fields []map[string]any) {
	fields = make([]map[string]any, 0)
	instanceType := reflect.TypeOf(role)

	// Todo : convert build to recursive for mode dynamic
	for i := 0; i < instanceType.NumField(); i++ {
		newField := make(map[string]any)
		field := instanceType.Field(i)
		if field.Name == "Password" {
			continue
		}
		fieldValue := reflect.ValueOf(role).Field(i)

		newField["Name"] = field.Name
		newField["Value"] = getRoleValue(field, fieldValue)
		fields = append(fields, newField)
	}

	return fields
}

func getRoleValue(field reflect.StructField, value reflect.Value) string {
	switch field.Type.Kind() {
	case reflect.String:
		return fmt.Sprintf("%q", value.String())
	case reflect.Ptr:
		if !value.IsNil() {
			return value.Elem().String()
		} else {
			return "nil"
		}
	case reflect.Slice:
		if field.Type.Elem().Kind() == reflect.String {
			return generateArrayDeclaration(value, false)
		}
	case reflect.Map:
		if len(value.MapKeys()) == 0 {
			return "map[string]any{}"
		}
		if mapValues, ok := value.Interface().(map[string]any); ok {
			mapValueStr := generateMapDeclaration(mapValues)
			return mapValueStr
		} else {
			return "map[string]any{}"
		}
	case reflect.Struct:
		return generateStructDataType(value)
	case reflect.Interface:
		rv := reflect.ValueOf(value.Interface())
		if rv.Kind() == reflect.Map {
			if len(rv.MapKeys()) == 0 {
				return "map[string]any{}"
			} else {
				return generateMapDeclarationFromValue(rv)
			}

		} else if rv.Kind() == reflect.Slice {
			return generateArrayDeclaration(rv, false)
		} else if !value.IsNil() {
			return fmt.Sprintf("%v", value.Interface())
		} else {
			return "nil"
		}
	}

	return fmt.Sprintf("%v", value.Interface())
}
