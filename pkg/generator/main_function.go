package generator

import (
	"fmt"
	"path/filepath"

	"github.com/sev-2/raiden"
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
	// load configuretion
	config := raiden.LoadConfig(nil)

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

func GenerateMainFunction(config *raiden.Config, generateFn GenerateFn) error {
	// make sure all folder exist
	cmdFolderPath := filepath.Join(config.ProjectName, "cmd")
	if exist := utils.IsFolderExists(cmdFolderPath); !exist {
		if err := utils.CreateFolder(cmdFolderPath); err != nil {
			return err
		}
	}

	mainFunctionDir := fmt.Sprintf(MainFunctionDirTemplate, config.ProjectName)
	folderPath := filepath.Join(config.ProjectName, mainFunctionDir)
	if err := utils.CreateFolder(folderPath); err != nil {
		return err
	}

	// set file path
	filePath := filepath.Join(folderPath, fmt.Sprintf("%s.%s", config.ProjectName, "go"))
	absolutePath, err := utils.GetAbsolutePath(filePath)
	if err != nil {
		return err
	}

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
	input := GenerateInput{
		BindData:     data,
		Template:     MainFunctionTemplate,
		TemplateName: "mainFunctionTemplate",
		OutputPath:   absolutePath,
	}

	return generateFn(input)
}
