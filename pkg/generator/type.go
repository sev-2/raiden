package generator

import (
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/hashicorp/go-hclog"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/utils"
)

var TypeLogger hclog.Logger = logger.HcLog().Named("generator.type")

// ----- Define type, variable and constant -----
type GenerateTypeData struct {
	Imports    []string
	Package    string
	Name       string
	Schema     string
	Format     string
	Enums      string
	Attributes string
	Comment    string
}

const (
	TypeDir      = "internal/types"
	TypeTemplate = `package {{ .Package }}
{{- if gt (len .Imports) 0 }}

import (
{{- range .Imports}}
	{{.}}
{{- end}}
)
{{- end }}

type {{ .Name | ToGoIdentifier }} struct {
	raiden.TypeBase
}

func (t *{{ .Name | ToGoIdentifier }}) Name() string {
	return "{{ .Name }}"
}
{{- if ne .Schema "public" }}

func (r *{{.Name | ToGoIdentifier }}) Schema() string {
	return "{{ .Schema }}"
}

{{- end }}

func (r *{{.Name | ToGoIdentifier }}) Format() string {
	return "{{ .Format }}"
}

func (r *{{.Name | ToGoIdentifier }}) Enums() []string {
	return {{ .Enums }}
}

func (r *{{.Name | ToGoIdentifier }}) Comment() *string {
	return {{ .Comment }}
}

`
)

func GenerateTypes(basePath string, types []objects.Type, generateFn GenerateFn) (err error) {
	folderPath := filepath.Join(basePath, TypeDir)
	TypeLogger.Trace("create types folder if not exist", folderPath)
	if exist := utils.IsFolderExists(folderPath); !exist {
		if err := utils.CreateFolder(folderPath); err != nil {
			return err
		}
	}

	for _, v := range types {
		if err := GenerateType(folderPath, v, generateFn); err != nil {
			return err
		}
	}

	return nil
}

func GenerateType(folderPath string, t objects.Type, generateFn GenerateFn) error {
	// define binding func
	funcMaps := []template.FuncMap{
		{"ToGoIdentifier": utils.SnakeCaseToPascalCase},
	}

	// define file path
	filePath := filepath.Join(folderPath, fmt.Sprintf("%s.%s", t.Name, "go"))

	// set imports path
	var imports []string
	raidenPath := fmt.Sprintf("%q", "github.com/sev-2/raiden")
	imports = append(imports, raidenPath)

	// execute the template and write to the file
	data := GenerateTypeData{
		Package:    "types",
		Imports:    imports,
		Name:       t.Name,
		Schema:     t.Schema,
		Format:     t.Format,
		Enums:      "[]string{}",
		Attributes: "[]string{}",
		Comment:    "nil",
	}

	if len(t.Enums) > 0 {
		enums := []string{}
		for _, e := range t.Enums {
			enums = append(enums, fmt.Sprintf("%q", e))
		}
		data.Enums = fmt.Sprintf("[]string{%s}", strings.Join(enums, ","))
	}

	if len(t.Attributes) > 0 {
		attributes := []string{}
		for _, e := range t.Attributes {
			attributes = append(attributes, fmt.Sprintf("%q", e))
		}
		data.Attributes = fmt.Sprintf("[]string{%s}", strings.Join(attributes, ","))
	}

	if t.Comment != nil {
		data.Comment = *t.Comment
	}

	// set input
	input := GenerateInput{
		BindData:     data,
		Template:     TypeTemplate,
		TemplateName: "typeTemplate",
		OutputPath:   filePath,
		FuncMap:      funcMaps,
	}

	TypeLogger.Debug("generate type", "path", input.OutputPath)
	return generateFn(input, nil)
}
