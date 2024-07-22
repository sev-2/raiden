package utils_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/sev-2/raiden/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestCreateFile(t *testing.T) {
	tmpDir := os.TempDir()
	filePath := filepath.Join(tmpDir, "test_file.txt")

	// Clean up before and after the test
	if utils.IsFileExists(filePath) {
		os.Remove(filePath)
	}
	defer os.Remove(filePath)

	// Test creating a new file
	file, err := utils.CreateFile(filePath, false)
	assert.NoError(t, err)
	assert.FileExists(t, filePath)
	file.Close()

	// Test creating an existing file with deleteIfExist = true
	file, err = utils.CreateFile(filePath, true)
	assert.NoError(t, err)
	assert.FileExists(t, filePath)
	file.Close()

	// Test creating an existing file with deleteIfExist = false
	file, err = utils.CreateFile(filePath, false)
	assert.NoError(t, err)
	assert.FileExists(t, filePath)
	file.Close()
}

func TestCopyFile(t *testing.T) {
	srcFile, err := ioutil.TempFile("", "src_file.txt")
	assert.NoError(t, err)
	defer os.Remove(srcFile.Name())

	_, err = srcFile.WriteString("Hello, World!")
	assert.NoError(t, err)
	srcFile.Close()

	dstFile, err := ioutil.TempFile("", "dst_file.txt")
	assert.NoError(t, err)
	dstFile.Close()
	defer os.Remove(dstFile.Name())

	err = utils.CopyFile(srcFile.Name(), dstFile.Name())
	assert.NoError(t, err)

	srcContent, err := ioutil.ReadFile(srcFile.Name())
	assert.NoError(t, err)

	dstContent, err := ioutil.ReadFile(dstFile.Name())
	assert.NoError(t, err)

	assert.Equal(t, srcContent, dstContent)
}

func TestDeleteFile(t *testing.T) {
	file, err := ioutil.TempFile("", "delete_file.txt")
	assert.NoError(t, err)
	filePath := file.Name()
	file.Close()

	err = utils.DeleteFile(filePath)
	assert.NoError(t, err)
	assert.False(t, utils.IsFileExists(filePath))
}

func TestIsFileExists(t *testing.T) {
	file, err := ioutil.TempFile("", "exists_file.txt")
	assert.NoError(t, err)
	filePath := file.Name()
	file.Close()
	defer os.Remove(filePath)

	assert.True(t, utils.IsFileExists(filePath))
	assert.False(t, utils.IsFileExists("/non_existing_file.txt"))
}
