package generator

import (
	"fmt"
	"path/filepath"

	"github.com/hashicorp/go-hclog"
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/utils"
)

var MainLogger hclog.Logger = logger.HcLog().Named("generator.main")

// ----- Define type, variable and constant -----
type GenerateMainFunctionData struct {
	Package string
	Imports []string
}

const (
	MainFunctionDirTemplate = "/cmd/%s"
	MainFunctionTemplate    = `package {{ .Package }}
{{- if gt (len .Imports) 0 }}

import (
{{- range .Imports}}
	{{.}}
{{- end}}
)
{{- end }}

func main() {
	// load configuration
	config, err := raiden.LoadConfig(nil)
	if err != nil {
		raiden.Error("load configuration",err.Error())
	}

	// Setup server
	server := raiden.NewServer(config)

	// register route
	bootstrap.RegisterRoute(server)

	// run server
	server.Run()
}
`
)

// ----- Generate main function -----

func GenerateMainFunction(basePath string, config *raiden.Config, generateFn GenerateFn) error {
	// make sure all folder exist
	cmdFolderPath := filepath.Join(basePath, "cmd")
	MainLogger.Trace("create cmd folder if not exist", "path", cmdFolderPath)
	if exist := utils.IsFolderExists(cmdFolderPath); !exist {
		if err := utils.CreateFolder(cmdFolderPath); err != nil {
			return err
		}
	}

	mainFunctionDir := fmt.Sprintf(MainFunctionDirTemplate, config.ProjectName)
	mainFunctionPath := filepath.Join(basePath, mainFunctionDir)
	MainLogger.Trace("create main folder if not exist", "path", mainFunctionPath)
	if exist := utils.IsFolderExists(mainFunctionPath); !exist {
		if err := utils.CreateFolder(mainFunctionPath); err != nil {
			return err
		}
	}

	// set file path
	filePath := filepath.Join(mainFunctionPath, fmt.Sprintf("%s.%s", utils.ToSnakeCase(config.ProjectName), "go"))

	// setup import path
	importPaths := []string{
		fmt.Sprintf("%q", "github.com/sev-2/raiden"),
	}
	routeImportPath := fmt.Sprintf("\"%s/internal/bootstrap\"", utils.ToGoModuleName(config.ProjectName))
	importPaths = append(importPaths, routeImportPath)
	data := GenerateMainFunctionData{
		Package: "main",
		Imports: importPaths,
	}

	// setup generate input param
	input := GenerateInput{
		BindData:     data,
		Template:     MainFunctionTemplate,
		TemplateName: "mainFunctionTemplate",
		OutputPath:   filePath,
	}

	MainLogger.Debug("generate main function", "path", input.OutputPath)
	return generateFn(input, nil)
}
