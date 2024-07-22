package utils_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sev-2/raiden/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestIsFolderExists(t *testing.T) {
	// Test with a non-existing folder
	nonExistentFolder := filepath.Join(os.TempDir(), "non_existent_folder")
	assert.False(t, utils.IsFolderExists(nonExistentFolder))

	// Test with an existing folder
	existingFolder := os.TempDir()
	assert.True(t, utils.IsFolderExists(existingFolder))
}

func TestGetCurrentDirectory(t *testing.T) {
	// Get the current working directory using os package
	expectedDir, err := os.Getwd()
	assert.NoError(t, err)

	// Get the current working directory using the utility function
	currentDir, err := utils.GetCurrentDirectory()
	assert.NoError(t, err)
	assert.Equal(t, expectedDir, currentDir)
}

func TestCreateFolder(t *testing.T) {
	// Create a temporary folder path
	tempFolder := filepath.Join(os.TempDir(), "test_folder")

	// Clean up before and after the test
	if utils.IsFolderExists(tempFolder) {
		os.RemoveAll(tempFolder)
	}
	defer os.RemoveAll(tempFolder)

	// Create the folder using the utility function
	err := utils.CreateFolder(tempFolder)
	assert.NoError(t, err)
	assert.True(t, utils.IsFolderExists(tempFolder))
}

func TestDeleteFolder(t *testing.T) {
	// Create a temporary folder
	tempFolder := filepath.Join(os.TempDir(), "test_folder")
	err := os.Mkdir(tempFolder, os.ModePerm)
	assert.NoError(t, err)
	assert.True(t, utils.IsFolderExists(tempFolder))

	// Delete the folder using the utility function
	err = utils.DeleteFolder(tempFolder)
	assert.NoError(t, err)
	assert.False(t, utils.IsFolderExists(tempFolder))
}

func TestGetAbsolutePath(t *testing.T) {
	// Get the current working directory using os package
	currentDir, err := os.Getwd()
	assert.NoError(t, err)

	// Define a relative path
	relativePath := "test_path"

	// Get the absolute path using the utility function
	absolutePath, err := utils.GetAbsolutePath(relativePath)
	assert.NoError(t, err)

	// Expected absolute path
	expectedAbsolutePath := filepath.Join(currentDir, relativePath)
	assert.Equal(t, expectedAbsolutePath, absolutePath)
}
