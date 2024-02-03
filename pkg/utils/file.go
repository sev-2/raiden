package utils

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sev-2/raiden/pkg/logger"
)

func CreateFile(fullPath string, deleteIfExist bool) (*os.File, error) {
	// configure file name
	folderPath := filepath.Dir(fullPath)
	fullFileName := filepath.Base(fullPath)

	fileExtension := filepath.Ext(fullFileName)
	fileName := fullFileName[:len(fullFileName)-len(fileExtension)]

	fileNameSnakeCase := ToSnakeCase(fileName)
	file := filepath.Join(folderPath, fmt.Sprintf("%s%s", fileNameSnakeCase, fileExtension))

	// check if file exist and than deleted
	if deleteIfExist && IsFileExists(file) {
		err := os.Remove(file)
		if err != nil {
			logger.Debugf("error delete file : %s", file)
		}
	}

	return os.Create(file)
}

func DeleteFile(filePath string) error {
	return os.Remove(filePath)
}

func IsFileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}
