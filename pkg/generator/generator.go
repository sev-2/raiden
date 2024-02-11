package generator

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"

	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/utils"
)

// ----- Define type, variable and constant -----
type GenerateInput struct {
	BindData     any
	Template     string
	TemplateName string
	OutputPath   string
	FuncMap      []template.FuncMap
}

type GenerateFn func(input GenerateInput, writer io.Writer) error

// ----- Generate functionality  -----

func DefaultWriter(filePath string) (*os.File, error) {
	file, err := utils.CreateFile(filePath, true)
	if err != nil {
		return nil, fmt.Errorf("failed create file %s : %v", filePath, err)
	}

	return file, nil
}

func Generate(input GenerateInput, writer io.Writer) error {

	// set default writer
	if writer == nil {
		file, err := DefaultWriter(input.OutputPath)
		if err != nil {
			return err
		}
		defer file.Close()
		writer = file
	}

	tmplInstance := template.New(input.TemplateName)
	for _, tm := range input.FuncMap {
		tmplInstance.Funcs(tm)
	}

	tmpl, err := tmplInstance.Parse(input.Template)
	if err != nil {
		return fmt.Errorf("error parsing : %v", err)
	}

	return tmpl.Execute(writer, input.BindData)
}

func CreateInternalFolder(basePath string) (err error) {
	internalFolderPath := filepath.Join(basePath, "internal")
	logger.Debugf("CreateInternalFolder - create `%s` folder if not exist", internalFolderPath)
	if exist := utils.IsFolderExists(internalFolderPath); !exist {
		if err := utils.CreateFolder(internalFolderPath); err != nil {
			return err
		}
	}
	return nil
}

func generateArrayDeclaration(value reflect.Value, withoutQuote bool) string {
	var arrayValues []string
	for i := 0; i < value.Len(); i++ {
		if withoutQuote {
			arrayValues = append(arrayValues, fmt.Sprintf("%s", value.Index(i).Interface()))
		} else {
			arrayValues = append(arrayValues, fmt.Sprintf("%q", value.Index(i).Interface()))
		}
	}
	return "[]string{" + strings.Join(arrayValues, ", ") + "}"
}

func generateMapDeclarationFromValue(value reflect.Value) string {
	// Start the map declaration
	var resultArr []string
	for _, key := range value.MapKeys() {
		valueStr := valueToString(value.Interface())
		resultArr = append(resultArr, fmt.Sprintf(`%q: %s,`, key, valueStr))
	}

	return "map[string]any{" + strings.Join(resultArr, ", ") + "}"
}

func generateMapDeclaration(mapData map[string]any) string {
	// Start the map declaration
	var resultArr []string
	for key, value := range mapData {
		valueStr := valueToString(value)
		resultArr = append(resultArr, fmt.Sprintf(`%q: %s`, key, valueStr))
	}

	return "map[string]any{" + strings.Join(resultArr, ",") + "}"
}

func generateStructDataType(value reflect.Value) string {
	return fmt.Sprintf("%v", value.Interface())
}

func valueToString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return fmt.Sprintf("%q", v)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%v", v)
	case float32, float64:
		return fmt.Sprintf("%v", v)
	case bool:
		return fmt.Sprintf("%t", v)
	default:
		return fmt.Sprintf(`"%v"`, v) // Default: use fmt.Sprint for other types
	}
}
