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

func formatArrayDataTypes(value reflect.Value) string {
	var arrayValues []string
	for i := 0; i < value.Len(); i++ {
		arrayValues = append(arrayValues, fmt.Sprintf("%q", value.Index(i).Interface()))
	}
	return "[]string{" + strings.Join(arrayValues, ", ") + "}"
}

func formatCustomStructDataType(value reflect.Value) string {
	return fmt.Sprintf("%v", value.Interface())
}
