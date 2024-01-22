package generator

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/utils"
)

var ControllerDir = "controllers"
var helloWordControllerTemplate = `package controllers

import (
	"github.com/sev-2/raiden"
)

// @type function
// @route /hello
func HelloWordHandler(ctx *raiden.Context) raiden.Presenter {
	response := map[string]any{
		"message": "hello word",
	}
	return ctx.SendData(response)
}
`

func GenerateHelloWordController(projectName string) (err error) {
	folderPath := filepath.Join(projectName, ControllerDir)
	err = utils.CreateFolder(folderPath)
	if err != nil {
		return
	}

	tmpl, err := template.New("helloWordControllerTemplate").Parse(helloWordControllerTemplate)
	if err != nil {
		return fmt.Errorf("error parsing template : %v", err)
	}

	// Create or open the output file
	file, err := createFile(getAbsolutePath(folderPath), "hello", "go")
	if err != nil {
		return fmt.Errorf("failed create file %s : %v", "hello", err)
	}
	defer file.Close()

	// Execute the template and write to the file
	err = tmpl.Execute(file, nil)
	if err != nil {
		return fmt.Errorf("error executing template: %v", err)
	}

	return nil
}

// Scan controller folder
func MapGetControllers(controllerDirPath string) (controller map[string]any, err error) {
	if controllerDirPath == "" {
		controllerDirPath = ControllerDir
	}

	controller = make(map[string]any)
	err = filepath.Walk(controllerDirPath, attachControllerToMap(controller))
	return
}

func attachControllerToMap(mapController map[string]any) filepath.WalkFunc {
	return func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".go") {
			handlerComments := raiden.GetHandlerComment(path)
			for _, hc := range handlerComments {
				mapController[hc["fn"]] = true
			}
		}
		return nil
	}
}
