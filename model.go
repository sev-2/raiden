package raiden

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/sev-2/raiden/pkg/utils"
)

var modelDir = "models"

func GetAllModelName(modelPath string) (modelNames []string, err error) {
	currentDir, errCurrentDir := utils.GetCurrentDirectory()
	if errCurrentDir != nil {
		return modelNames, errCurrentDir
	}

	modelDirPath := filepath.Join(currentDir, modelDir)
	err = filepath.Walk(modelDirPath, walkScanModelName(&modelNames))
	return
}

func walkScanModelName(modelNames *[]string) filepath.WalkFunc {
	return func(path string, info fs.FileInfo, err error) error {
		tmpModelNames := *modelNames
		if err != nil {
			return err
		}

		if strings.HasSuffix(path, ".go") {
			foundedModelNames, err := scanModelName(path)
			if err != nil {
				return err
			}
			tmpModelNames = append(tmpModelNames, foundedModelNames...)
		}

		modelNames = &tmpModelNames
		return nil
	}
}

func scanModelName(file string) (modelNames []string, err error) {
	fSet := token.NewFileSet()
	parsedFile, err := parser.ParseFile(fSet, file, nil, parser.ParseComments)
	if err != nil {
		return
	}

	ast.Inspect(parsedFile, func(node ast.Node) bool {
		if typeSpec, ok := node.(*ast.TypeSpec); ok {
			if _, isStruct := typeSpec.Type.(*ast.StructType); isStruct {
				modelNames = append(modelNames, typeSpec.Name.Name)
			}
		}
		return true
	})
	return
}
