package generator

import (
	"fmt"
	"path/filepath"
	"reflect"
	"text/template"

	"github.com/sev-2/raiden/pkg/supabase"
	"github.com/sev-2/raiden/pkg/utils"
)

var roleDir = "internal/roles"
var roleInstanceTemplate = `package roles

import (
	"github.com/sev-2/raiden/pkg/postgres"
)

var {{ .RoleName | ToGoIdentifier }} = &postgres.Role{
{{- range $i, $field := .Fields }}
	{{ .Name }} : {{ .Value }},
{{- end }}
}
`

func GenerateRoles(projectName string, roles []supabase.Role) (err error) {
	internalFolderPath := filepath.Join(projectName, "internal")
	if exist := utils.IsFolderExists(internalFolderPath); !exist {
		if err := utils.CreateFolder(internalFolderPath); err != nil {
			return err
		}
	}

	err = utils.CreateFolder(filepath.Join(projectName, roleDir))
	if err != nil {
		return err
	}

	for _, v := range roles {
		if err := GenerateRole(projectName, v); err != nil {
			return err
		}
	}

	return nil
}

func GenerateRole(projectName string, role supabase.Role) error {
	tmpl, err := template.New("roleInstanceTemplate").
		Funcs(template.FuncMap{"ToGoIdentifier": utils.SnakeCaseToPascalCase}).
		Parse(roleInstanceTemplate)
	if err != nil {
		return fmt.Errorf("error parsing template : %v", err)
	}

	fields := make([]map[string]any, 0)
	instanceType := reflect.TypeOf(role)

	// Todo : convert build to recursive
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
	// Create or open the output file
	folderPath := fmt.Sprintf("%s/%s", projectName, roleDir)
	file, err := createFile(getAbsolutePath(folderPath), role.Name, "go")
	if err != nil {
		return fmt.Errorf("failed create file %s : %v", role.Name, err)
	}
	defer file.Close()

	// Execute the template and write to the file
	err = tmpl.Execute(file, map[string]any{
		"RoleName":   role.Name,
		"Fields":     fields,
		"TotalField": len(fields),
	})

	if err != nil {
		return fmt.Errorf("error executing template: %v", err)
	}

	return nil
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
			return generateArrayDeclaration(value)
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
		return formatCustomStructDataType(value)
	case reflect.Interface:
		rv := reflect.ValueOf(value.Interface())
		if rv.Kind() == reflect.Map {
			if len(rv.MapKeys()) == 0 {
				return "map[string]any{}"
			} else {
				return generateMapDeclarationFromValue(rv)
			}

		} else if rv.Kind() == reflect.Slice {
			return generateArrayDeclaration(rv)
		} else if !value.IsNil() {
			return fmt.Sprintf("%v", value.Interface())
		} else {
			return "nil"
		}
	}

	return fmt.Sprintf("%v", value.Interface())
}
