package state_test

import (
	"testing"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
)

type MockType struct {
	raiden.TypeBase
}

func (r *MockType) Name() string {
	return "type_1"
}

func (r *MockType) Schema() string {
	return raiden.DefaultTypeSchema
}

func (*MockType) Format() string {
	return ""
}

func (*MockType) Enums() []string {
	return []string{"type_1", "type_2"}
}

func (*MockType) Attributes() []objects.TypeAttribute {
	return []objects.TypeAttribute{}
}

func (*MockType) Comment() *string {
	return nil
}

func TestExtractType(t *testing.T) {
	typeStates := []state.TypeState{
		{Type: objects.Type{Name: "existing_type"}},
		{Type: objects.Type{Name: "test_type"}},
	}

	appTypes := []raiden.Type{
		&MockType{},
	}

	result, err := state.ExtractType(typeStates, appTypes)
	assert.NoError(t, err)
	assert.Len(t, result.Existing, 0)
	assert.Len(t, result.New, 1)
	assert.Len(t, result.Delete, 2)
}

func TestBindToSupabaseType(t *testing.T) {
	dataType := MockType{}
	r := objects.Type{}

	state.BindToSupabaseType(&r, &dataType)
	assert.Equal(t, "type_1", r.Name)
	assert.Equal(t, "public", r.Schema)
	assert.Equal(t, "", r.Format)
	assert.Equal(t, 2, len(r.Enums))
	assert.Equal(t, 0, len(r.Attributes))
	assert.Nil(t, r.Comment)
}

func TestBuildTypeFromState(t *testing.T) {
	rs := state.TypeState{
		Type: objects.Type{
			Name:       "type_1",
			Schema:     "test",
			Format:     "type_1",
			Enums:      []string{"type_1", "type_2"},
			Attributes: []objects.TypeAttribute{},
			Comment:    nil,
		},
	}
	dataType := &MockType{}

	r := state.BuildTypeFromState(rs, dataType)
	assert.Equal(t, "type_1", r.Name)
	assert.Equal(t, "public", r.Schema)
	assert.Equal(t, "", r.Format)
	assert.Equal(t, 2, len(r.Enums))
	assert.Equal(t, 0, len(r.Attributes))
	assert.Nil(t, r.Comment)
}

func TestExtractTypeResult_ToDeleteFlatMap(t *testing.T) {
	extractTypeResult := state.ExtractTypeResult{
		Delete: []objects.Type{
			{Name: "type1"},
			{Name: "type2"},
		},
	}

	mapData := extractTypeResult.ToDeleteFlatMap()
	assert.Len(t, mapData, 2)
	assert.Contains(t, mapData, "type1")
	assert.Contains(t, mapData, "type2")
	assert.Equal(t, "type1", mapData["type1"].Name)
	assert.Equal(t, "type2", mapData["type2"].Name)
}

func TestExtractTypeResult_ToMap(t *testing.T) {
	extractTypeResult := state.ExtractTypeResult{
		Existing: []objects.Type{
			{Name: "type1"},
			{Name: "type2"},
		},
	}

	mapData := extractTypeResult.ToMap()
	assert.Len(t, mapData, 2)
	assert.Contains(t, mapData, "type1")
	assert.Contains(t, mapData, "type2")
	assert.Equal(t, "type1", mapData["type1"].Name)
	assert.Equal(t, "type2", mapData["type2"].Name)
}
