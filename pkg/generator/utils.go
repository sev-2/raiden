package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/utils"
)

func createFile(folderPath, fileName string, extension string) (*os.File, error) {
	snakeCaseFile := fmt.Sprintf("%s.%s", utils.ToSnakeCase(fileName), extension)
	file := filepath.Join(folderPath, snakeCaseFile)

	// check file exist than delete
	if _, err := os.Stat(file); err == nil {
		err := os.Remove(file)
		if err != nil {
			logger.Debugf("error delete file : %s", file)
		}
	}

	return os.Create(file)
}

func getAbsolutePath(folderPath string) string {
	currDir, err := utils.GetCurrentDirectory()
	if err != nil {
		logger.Panic(err)
	}

	return filepath.Join(currDir, folderPath)
}

func generateArrayDeclaration(value reflect.Value) string {
	var arrayValues []string
	for i := 0; i < value.Len(); i++ {
		arrayValues = append(arrayValues, fmt.Sprintf("%q", value.Index(i).Interface()))
	}
	return "[]string{" + strings.Join(arrayValues, ", ") + "}"
}

func formatCustomStructDataType(value reflect.Value) string {
	return fmt.Sprintf("%v", value.Interface())
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
