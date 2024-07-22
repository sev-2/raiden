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

func TestIsStringContainSpace(t *testing.T) {
	assert.True(t, utils.IsStringContainSpace("Hello World"))
	assert.False(t, utils.IsStringContainSpace("HelloWorld"))
}

func TestToSnakeCase(t *testing.T) {
	assert.Equal(t, "hello_world", utils.ToSnakeCase("HelloWorld"))
	assert.Equal(t, "hello_world", utils.ToSnakeCase("hello-world"))
}

func TestSnakeCaseToPascalCase(t *testing.T) {
	assert.Equal(t, "HelloWorld", utils.SnakeCaseToPascalCase("hello_world"))
	assert.Equal(t, "HelloWorld", utils.SnakeCaseToPascalCase("Hello_World"))
}

func TestMatchReplacer(t *testing.T) {
	query := "SELECT * FROM users WHERE name = :name AND age > :age"
	paramKey := ":name"
	replacement := "'John Doe'"
	expected := "SELECT * FROM users WHERE name = 'John Doe' AND age > :age"
	assert.Equal(t, expected, utils.MatchReplacer(query, paramKey, replacement))
}

func TestCleanUpString(t *testing.T) {
	input := "Hello,\tWorld!\nThis is a test."
	expected := "Hello, World! This is a test."
	assert.Equal(t, expected, utils.CleanUpString(input))
}

func TestHashString(t *testing.T) {
	query := "SELECT * FROM users"
	hashed := utils.HashString(query)
	assert.Equal(t, 64, len(hashed)) // SHA256 hash length
}

func TestToCamelCase(t *testing.T) {
	assert.Equal(t, "hello_world", utils.ToCamelCase("HelloWorld"))
	assert.Equal(t, "my_variable_name", utils.ToCamelCase("MyVariableName"))
}

func TestParseBool(t *testing.T) {
	assert.True(t, utils.ParseBool("true"))
	assert.False(t, utils.ParseBool("false"))
}

func TestConvertAllToString(t *testing.T) {
	assert.Equal(t, "42", utils.ConvertAllToString(42))
	assert.Equal(t, "3.14", utils.ConvertAllToString(3.14))
	assert.Equal(t, "hello", utils.ConvertAllToString("hello"))
	assert.Equal(t, "true", utils.ConvertAllToString(true))
	assert.Equal(t, "[1 2 3]", utils.ConvertAllToString([]int{1, 2, 3}))
	assert.Equal(t, "[a b c]", utils.ConvertAllToString([]string{"a", "b", "c"}))
	assert.Equal(t, "map[a:1 b:2]", utils.ConvertAllToString(map[string]int{"a": 1, "b": 2}))
	assert.Equal(t, "nil", utils.ConvertAllToString(nil))
	assert.Equal(t, "unknown type", utils.ConvertAllToString(struct{}{}))
}

func TestCleanDoubleColonPattern(t *testing.T) {
	input := "text::timestamp"
	expected := "text"
	assert.Equal(t, expected, utils.CleanDoubleColonPattern(input))
}
