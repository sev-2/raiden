package generator

import (
	"fmt"
	"path/filepath"
	"reflect"
	"text/template"

	"github.com/sev-2/raiden/pkg/supabase"
	"github.com/sev-2/raiden/pkg/utils"
)

var roleDir = "roles"
var roleInstanceTemplate = `
package roles

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

	for i := 0; i < instanceType.NumField(); i++ {
		field := instanceType.Field(i)
		if field.Name == "Password" {
			continue
		}
		fieldValue := reflect.ValueOf(role).Field(i)

		newField := make(map[string]any)
		newField["Name"] = field.Name

		switch field.Type.Kind() {
		case reflect.String:
			newField["Value"] = fmt.Sprintf("%q", fieldValue.String())
		case reflect.Ptr:
			if !fieldValue.IsNil() {
				newField["Value"] = fieldValue.Elem()
			} else {
				newField["Value"] = "nil"
			}
		case reflect.Slice:
			if field.Type.Elem().Kind() == reflect.String {
				newField["Value"] = formatArrayDataTypes(fieldValue)
			}
		case reflect.Struct:
			newField["Value"] = formatCustomStructDataType(fieldValue)
		default:
			newField["Value"] = fmt.Sprintf("%v", fieldValue.Interface())
		}

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
