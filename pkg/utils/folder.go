package utils

import (
	"os"
	"path/filepath"
)

func IsFolderExists(folderPath string) bool {
	_, err := os.Stat(folderPath)
	if err == nil {
		return true
	}

	if os.IsNotExist(err) {
		return false
	}

	return false
}

func GetCurrentDirectory() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return dir, nil
}

func CreateFolder(folderName string) error {
	currentDir, err := GetCurrentDirectory()
	if err != nil {
		return err
	}

	folderPath := filepath.Join(currentDir, folderName)
	if IsFolderExists(folderPath) {
		return nil
	}

	err = os.Mkdir(folderPath, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

func DeleteFolder(folderName string) error {
	currentDir, err := GetCurrentDirectory()
	if err != nil {
		return err
	}

	folderPath := filepath.Join(currentDir, folderName)
	if !IsFolderExists(folderPath) {
		return nil
	}

	err = os.RemoveAll(folderPath)
	if err != nil {
		return err
	}
	return nil
}
