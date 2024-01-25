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

var ControllerDir = "internal/controllers"
var helloWordControllerTemplate = `package controllers

import (
	"fmt"
	
	"github.com/sev-2/raiden"
)

type HelloWordRequest struct {
	Name    string ` + "`path:\"name\" validate:\"required\"`" + ` 
	Type    string ` + "`query:\"type\" validate:\"required\"`" + `
	Message string ` + "`json:\"message\" validate:\"required\"`" + `
}

type HelloWordResponse struct {
	Type    string ` + "`json:\"type\"`" + `
	Message string ` + "`json:\"message\"`" + `
}

// @type function
// @route /hello/{name}
func HelloWordHandler(ctx raiden.Context) raiden.Presenter {
	payload, err := raiden.UnmarshalRequestAndValidate[HelloWordRequest](ctx)
	if err != nil {
		return ctx.SendJsonError(err)
	}

	response := HelloWordResponse{
		Message: fmt.Sprintf("hello %s, %s", payload.Name, payload.Message),
		Type:    payload.Type,
	}

	return ctx.SendJson(response)
}
`

func GenerateHelloWordController(projectName string) (err error) {
	internalFolderPath := filepath.Join(projectName, "internal")
	if exist := utils.IsFolderExists(internalFolderPath); !exist {
		if err := utils.CreateFolder(internalFolderPath); err != nil {
			return err
		}
	}

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
