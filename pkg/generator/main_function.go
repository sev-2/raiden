package generator

import (
	"fmt"
	"path/filepath"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/utils"
)

// ----- Define type, var and constant -----
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
		raiden.Panic(err)
	}

	// Setup server
	server := raiden.NewServer(config)

	// register route
	router.Register(server)

	// run server
	server.Run()
}
`
)

// ----- Generate main function -----

func GenerateMainFunction(basePath string, config *raiden.Config, generateFn GenerateFn) error {
	// make sure all folder exist
	cmdFolderPath := filepath.Join(basePath, "cmd")
	logger.Debugf("GenerateMainFunction - create %s folder if not exist", cmdFolderPath)
	if exist := utils.IsFolderExists(cmdFolderPath); !exist {
		if err := utils.CreateFolder(cmdFolderPath); err != nil {
			return err
		}
	}

	mainFunctionDir := fmt.Sprintf(MainFunctionDirTemplate, config.ProjectName)
	mainFunctionPath := filepath.Join(basePath, mainFunctionDir)
	logger.Debugf("GenerateMainFunction - create %s folder if not exist", mainFunctionPath)
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
	routeImportPath := fmt.Sprintf("\"%s/internal/router\"", utils.ToGoModuleName(config.ProjectName))
	importPaths = append(importPaths, routeImportPath)
	data := GenerateMainFunctionData{
		Package: "main",
		Imports: importPaths,
	}

	// setup generate input param
	logger.Debug("GenerateMainFunction - create main function input")
	input := GenerateInput{
		BindData:     data,
		Template:     MainFunctionTemplate,
		TemplateName: "mainFunctionTemplate",
		OutputPath:   filePath,
	}

	logger.Debugf("GenerateMainFunction - generate main function to %s", input.OutputPath)
	return generateFn(input)
}
