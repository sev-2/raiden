package utils_test

import (
	"testing"

	"github.com/sev-2/raiden/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestToGoModuleCorrectWithPathPayload(t *testing.T) {
	folderPath := "Project/Internal/HelloController"

	result := utils.ToGoModuleName(folderPath)
	assert.Equal(t, "hellocontroller", result)
}
