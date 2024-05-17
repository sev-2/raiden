package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/sev-2/raiden/pkg/logger"
)

var FileLogger = logger.HcLog().Named("raiden.utils.file")

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
			FileLogger.Error("delete file", "file", file)
		}
	}

	return os.Create(file)
}

func CopyFile(sourcePath, targetPath string) error {
	// Open the source file for reading
	src, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("error opening source file: %s", err)
	}
	defer src.Close()

	// Create the destination file
	dst, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("error creating destination file: %s", err)
	}
	defer dst.Close()

	// Copy the contents of the source file to the destination file
	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("error copying file: %s", err)
	}

	return nil
}

func DeleteFile(filePath string) error {
	return os.Remove(filePath)
}

func IsFileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}
