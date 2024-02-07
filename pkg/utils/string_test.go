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

func TestRemoveParenthesesContent(t *testing.T) {
	testString := "varchar(20)"

	rs := utils.RemoveParenthesesContent(testString)
	assert.Equal(t, "varchar", rs)
}

func TestParseTag(t *testing.T) {
	rawTag := `config:"key1:value1;key2:value2" connectionLimit:"60" inheritRole:"true" isReplicationRole:"true" isSuperuser:"true"`
	mapTag := utils.ParseTag(rawTag)
	assert.Equal(t, "key1:value1;key2:value2", mapTag["config"])
	assert.Equal(t, "60", mapTag["connectionLimit"])
	assert.Equal(t, "true", mapTag["inheritRole"])
	assert.Equal(t, "true", mapTag["isReplicationRole"])
	assert.Equal(t, "true", mapTag["isSuperuser"])
}
