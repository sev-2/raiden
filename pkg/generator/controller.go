package generator

import (
	"fmt"
	"path/filepath"

	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/utils"
)

// ----- Define type, variable and constant -----
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

func (c *{{ .Name }}Controller) Get(ctx raiden.Context) error {
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
	logger.Debugf("GenerateController - generate controller to %s", input.OutputPath)
	return generateFn(input, nil)
}

// ----- Generate hello word -----
func GenerateHelloWordController(basePath string, generateFn GenerateFn) (err error) {
	controllerPath := filepath.Join(basePath, ControllerDir)
	logger.Debugf("GenerateHelloWordController - create %s folder if not exist", controllerPath)
	if exist := utils.IsFolderExists(controllerPath); !exist {
		if err := utils.CreateFolder(controllerPath); err != nil {
			return err
		}
	}
	return createHelloWordController(controllerPath, generateFn)
}

func createHelloWordController(controllerPath string, generateFn GenerateFn) error {
	// set file path
	filePath := filepath.Join(controllerPath, fmt.Sprintf("%s.go", "hello"))

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
		HttpTag:        "`path:\"/hello/{name}\" type:\"custom\"`",
		Imports:        imports,
		RequestFields:  requestFields,
		ResponseFields: responseField,
		DefaultAction:  "c.Result.Message = \"Welcome to raiden\"",
	}

	return GenerateController(filePath, data, generateFn)
}
