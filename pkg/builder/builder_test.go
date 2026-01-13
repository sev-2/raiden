package builder_test

import (
	"testing"

	"github.com/google/uuid"
	builder "github.com/sev-2/raiden/pkg/builder"
	"github.com/sev-2/raiden/pkg/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type sampleModel struct {
	Field  string    `json:"field"`
	Id     uuid.UUID `json:"id,omitempty" column:"name:id;type:uuid;primaryKey;nullable:false;default:gen_random_uuid()"`
	UserId uuid.UUID `json:"user_id,omitempty" column:"name:user_id;type:uuid;nullable:false"`
}

func (s *sampleModel) GetIdColName() string {
	return builder.ColOf(s, s.Id)
}

func TestColOfReturnsColumnName(t *testing.T) {
	m := sampleModel{}
	col := builder.ColOf(&m, &m.Field)
	require.Equal(t, "field", col)
}

func TestColOfUuuidReturnsColumnName(t *testing.T) {
	m := sampleModel{}
	require.Equal(t, "id", m.GetIdColName())
}

func TestColOfPanicsWhenModelNotStructPointer(t *testing.T) {
	field := "value"
	expected := "builder: model must resolve to struct (got *string)"
	require.PanicsWithValue(t, expected, func() {
		builder.ColOf(&field, &field)
	})
}

func TestColOfPanicsWhenFieldPointerNil(t *testing.T) {
	m := sampleModel{}
	require.PanicsWithValue(t, "builder: fieldPtr must be a pointer to a struct field (e.g., &m.Title)", func() {
		builder.ColOf(&m, (*string)(nil))
	})
}

func TestColOfWithStructValue(t *testing.T) {
	m := sampleModel{}
	col := builder.ColOf(m, &m.Field)
	require.Equal(t, "field", col)
}

func TestColOfWithEmbeddedStruct(t *testing.T) {
	type course struct {
		db.ModelBase
		Owner *string `column:"name:owner"`
	}

	c := course{}
	require.Equal(t, "owner", builder.ColOf(&c, &c.Owner))
}

// Test TableFromModel function
func TestTableFromModel(t *testing.T) {
	t.Run("with struct", func(t *testing.T) {
		type TestModel struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		}
		schema, table := builder.TableFromModel(TestModel{})
		assert.Equal(t, "public", schema)
		assert.Equal(t, "test_models", table)
	})

	t.Run("with pointer to struct", func(t *testing.T) {
		type TestModel struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		}
		schema, table := builder.TableFromModel(&TestModel{})
		assert.Equal(t, "public", schema)
		assert.Equal(t, "test_models", table)
	})

	t.Run("with non-struct", func(t *testing.T) {
		schema, table := builder.TableFromModel("string")
		assert.Equal(t, "public", schema)
		assert.Equal(t, "unknown", table)
	})

	t.Run("with schema tag", func(t *testing.T) {
		type TestModel struct {
			ID   int    `json:"id" schema:"custom_schema"`
			Name string `json:"name"`
		}
		schema, table := builder.TableFromModel(TestModel{})
		assert.Equal(t, "custom_schema", schema)
		assert.Equal(t, "test_models", table)
	})

	t.Run("with tableName tag", func(t *testing.T) {
		type TestModel struct {
			ID   int    `json:"id" tableName:"custom_table"`
			Name string `json:"name"`
		}
		schema, table := builder.TableFromModel(TestModel{})
		assert.Equal(t, "public", schema)
		assert.Equal(t, "custom_table", table)
	})

	t.Run("with both schema and tableName tags", func(t *testing.T) {
		type TestModel struct {
			ID   int    `json:"id" schema:"custom_schema" tableName:"custom_table"`
			Name string `json:"name"`
		}
		schema, table := builder.TableFromModel(TestModel{})
		assert.Equal(t, "custom_schema", schema)
		assert.Equal(t, "custom_table", table)
	})
}

// Test parseColumnName function - test through a specific struct that uses column tag
func TestParseColumnName(t *testing.T) {
	type TestModelWithColumnTag struct {
		CustomField string `column:"name:custom_name" json:"custom_field"`
	}

	m := TestModelWithColumnTag{}
	assert.Equal(t, "custom_name", builder.ColOf(&m, &m.CustomField))
}

// Test colNameFromPtr with various inputs to cover different code paths
func TestColNameFromPtrEdgeCases(t *testing.T) {
	t.Run("with model and field as nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "builder: model and fieldPtr are required", func() {
			builder.ColOf(nil, nil)
		})
	})

	t.Run("with fieldPtr as nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "builder: model and fieldPtr are required", func() {
			builder.ColOf(sampleModel{}, nil)
		})
	})

	t.Run("with non-struct model", func(t *testing.T) {
		model := "not a struct"
		field := "field"
		assert.PanicsWithValue(t, "builder: model must resolve to struct (got string)", func() {
			builder.ColOf(model, &field)
		})
	})
}

// Test pointer field
func TestPointerField(t *testing.T) {
	type TestModel struct {
		PtrField *string `json:"ptr_field"`
		Value    string  `json:"value"`
	}

	model := TestModel{}
	model.PtrField = new(string)
	*model.PtrField = "test"

	col := builder.ColOf(&model, &model.PtrField)
	assert.Equal(t, "ptr_field", col)

	col2 := builder.ColOf(&model, &model.Value)
	assert.Equal(t, "value", col2)
}

// Test fieldMatchScore function through ColOf (the functionality is internal)
func TestFieldMatchScore(t *testing.T) {
	// This function is tested through the ColOf function when it needs to score fields
	type TestModel struct {
		ColumnField string `column:"name:custom_col" json:"custom"`
	}

	m := TestModel{}
	col := builder.ColOf(&m, &m.ColumnField)
	assert.Equal(t, "custom_col", col) // This path exercises fieldMatchScore with column tag
}
