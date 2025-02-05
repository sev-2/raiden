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

func (*MockType) Attributes() []string {
	return []string{}
}

func (*MockType) Comment() *string {
	return nil
}

func TestExtractType(t *testing.T) {
	roleStates := []state.TypeState{
		{Type: objects.Type{Name: "existing_role"}},
		{Type: objects.Type{Name: "test_role"}},
	}

	appTypes := []raiden.Type{
		&MockType{},
	}

	result, err := state.ExtractType(roleStates, appTypes)
	assert.NoError(t, err)
	assert.Len(t, result.Existing, 0)
	assert.Len(t, result.New, 1)
	assert.Len(t, result.Delete, 2)
}

func TestBindToSupabaseType(t *testing.T) {
	role := MockType{}
	r := objects.Type{}

	state.BindToSupabaseType(&r, &role)
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
			Attributes: []string{},
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
			{Name: "role1"},
			{Name: "role2"},
		},
	}

	mapData := extractTypeResult.ToDeleteFlatMap()
	assert.Len(t, mapData, 2)
	assert.Contains(t, mapData, "role1")
	assert.Contains(t, mapData, "role2")
	assert.Equal(t, "role1", mapData["role1"].Name)
	assert.Equal(t, "role2", mapData["role2"].Name)
}
