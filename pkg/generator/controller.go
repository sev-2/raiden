package generator

import (
	"fmt"
	"path/filepath"

	"github.com/sev-2/raiden/pkg/utils"
)

type ControllerFieldAttribute struct {
	Field string
	Type  string
	Tag   string
}

type GenerateControllerData struct {
	DefaultAction  string
	HttpTag        string
	Imports        []string
	Name           string
	Package        string
	RequestFields  []ControllerFieldAttribute
	ResponseFields []ControllerFieldAttribute
}

const (
	ControllerDir      = "internal/controllers"
	ControllerTemplate = `package {{ .Package }}
{{- if gt (len .Imports) 0 }}

import (
{{- range .Imports}}
	{{.}}
{{- end}}
)
{{- end }}

type {{ .Name }}Request struct {
	{{- if not .RequestFields}} // Add your request data {{ else }}
	{{- range .RequestFields}}
	{{ .Field }} {{ .Type }} {{ .Tag }}
	{{- end }}
	{{- end }}
}

type {{ .Name }}Response struct {
	{{- if not .ResponseFields}} // Add your response data {{ else }}
	{{- range .ResponseFields}}
	{{ .Field }} {{ .Type }} {{ .Tag }}
	{{- end}}
	{{- end }}
}

type {{ .Name }}Controller struct {
	raiden.ControllerBase
	Http	string ` + "{{ .HttpTag }}" + `
	Payload	*{{ .Name }}Request
	Result	{{ .Name }}Response
}

func (c *Controller) Handler(ctx raiden.Context) raiden.Presenter {
	{{ .DefaultAction }}
	return ctx.SendJson(c.Result)
}
`
)

// ----- Generate controller -----
func GenerateController(file string, data GenerateControllerData, generateFn GenerateFn) error {
	input := GenerateInput{
		BindData:     data,
		Template:     ControllerTemplate,
		TemplateName: "controllerName",
		OutputPath:   file,
	}
	return generateFn(input)
}

// ----- Generate hello word -----
func GenerateHelloWordController(projectName string, generateFn GenerateFn) (err error) {
	internalFolderPath := filepath.Join(projectName, "internal")
	if exist := utils.IsFolderExists(internalFolderPath); !exist {
		if err := utils.CreateFolder(internalFolderPath); err != nil {
			return err
		}
	}

	controllerPath := filepath.Join(projectName, ControllerDir)
	if exist := utils.IsFolderExists(controllerPath); !exist {
		if err := utils.CreateFolder(controllerPath); err != nil {
			return err
		}
	}
	return createHelloWordController(controllerPath, generateFn)
}

func createHelloWordController(controllerPath string, generateFn GenerateFn) error {
	filePath := filepath.Join(controllerPath, fmt.Sprintf("%s.go", "hello"))
	absolutePath, err := utils.GetAbsolutePath(filePath)
	if err != nil {
		return err
	}

	// set imports path
	imports := []string{
		fmt.Sprintf("%q", "github.com/sev-2/raiden"),
	}

	// set passed parameter
	requestFields := []ControllerFieldAttribute{}
	responseField := []ControllerFieldAttribute{
		{
			Field: "Message",
			Type:  "string",
			Tag:   fmt.Sprintf("`json:%q`", "message"),
		},
	}

	data := GenerateControllerData{
		Name:           "HelloWord",
		Package:        "controllers",
		HttpTag:        "`verb:\"post\" path:\"/hello/{name}\" type:\"http-handler\"`",
		Imports:        imports,
		RequestFields:  requestFields,
		ResponseFields: responseField,
		DefaultAction:  "c.Result.Message = \"Welcome to raiden\"",
	}

	return GenerateController(absolutePath, data, generateFn)
}
