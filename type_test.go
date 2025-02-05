package raiden_test

import (
	"encoding/json"
	"testing"

	"github.com/sev-2/raiden"
	"github.com/stretchr/testify/assert"
)

const (
	DefaultTypeSchema = "public"
)

func TestTypeBase_Name(t *testing.T) {
	typeBase := raiden.TypeBase{}
	assert.Equal(t, "", typeBase.Name())
}

func TestTypeBase_Schema(t *testing.T) {
	typeBase := raiden.TypeBase{}
	assert.Equal(t, raiden.DefaultTypeSchema, typeBase.Schema())
}

func TestTypeBase_Format(t *testing.T) {
	typeBase := raiden.TypeBase{}
	assert.Equal(t, "", typeBase.Format())
}

func TestTypeBase_Enums(t *testing.T) {
	typeBase := raiden.TypeBase{}
	assert.Equal(t, 0, len(typeBase.Enums()))
}

func TestTypeBase_Attribute(t *testing.T) {
	typeBase := raiden.TypeBase{}
	assert.Equal(t, false, typeBase.Attributes())
}

func TestTypeBase_Comment(t *testing.T) {
	typeBase := raiden.TypeBase{}
	assert.Equal(t, nil, typeBase.Comment())
}

func TestTypeBase_Value(t *testing.T) {
	typeBase := raiden.TypeBase{
		Value: "test",
	}
	assert.Equal(t, "test", typeBase.String())
}

func TestTypeBase_SetValue(t *testing.T) {
	typeBase := raiden.TypeBase{}
	typeBase.SetValue("test")
	assert.Equal(t, "test", typeBase.String())
}

type MockType struct {
	Type raiden.TypeBase `json:"mock_type"`
}

func TestTypeBase_JsonUnmarshal(t *testing.T) {
	jsonStr := "{\"mock_type\": \"test_type\"}"

	var mockType MockType
	err := json.Unmarshal([]byte(jsonStr), &mockType)
	assert.NoError(t, err)
	assert.Equal(t, "test_type", mockType.Type.Value)
}

func TestTypeBase_JsonMarshall(t *testing.T) {
	jsonStr := "{\"mock_type\":\"test_type\"}"

	mockType := MockType{
		Type: raiden.TypeBase{
			Value: "test_type",
		},
	}
	byteData, err := json.Marshal(&mockType)
	assert.NoError(t, err)

	assert.Equal(t, jsonStr, string(byteData))
}
