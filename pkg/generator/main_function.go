package generator

import (
	"fmt"
	"path/filepath"
	"text/template"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/utils"
)

var mainFunctionDirTemplate = "/cmd/%s"
var mainFunctionTemplate = `package main
import (
	"github.com/sev-2/raiden"
)

func main() {
	config := raiden.LoadConfig(nil)

	// Setup server
	app := raiden.NewServer(config)
	app.Run()
}
`

func GenerateMainFunction(config raiden.Config) error {
	cmdFolderPath := filepath.Join(config.ProjectName, "cmd")
	if exist := utils.IsFolderExists(cmdFolderPath); !exist {
		if err := utils.CreateFolder(cmdFolderPath); err != nil {
			return err
		}
	}

	mainFunctionDir := fmt.Sprintf(mainFunctionDirTemplate, config.ProjectName)
	folderPath := filepath.Join(config.ProjectName, mainFunctionDir)
	err := utils.CreateFolder(folderPath)
	if err != nil {
		return err
	}

	tmpl, err := template.New("mainFunctionTemplate").Parse(mainFunctionTemplate)
	if err != nil {
		return fmt.Errorf("error parsing template : %v", err)
	}

	// Create or open the output file
	file, err := createFile(getAbsolutePath(folderPath), config.ProjectName, "go")
	if err != nil {
		return fmt.Errorf("failed create file %s : %v", config.ProjectName, err)
	}
	defer file.Close()

	// Execute the template and write to the file
	err = tmpl.Execute(file, config)
	if err != nil {
		return fmt.Errorf("error executing template: %v", err)
	}

	return nil
}
