package raiden_test

import (
	"testing"

	"github.com/sev-2/raiden"
	"github.com/stretchr/testify/assert"
)

// ---- Test Column Tag
func TestUnMarshallColumnTag(t *testing.T) {
	tag := "name:name;type:varchar(10);nullable;unique"

	column := raiden.UnmarshalColumnTag(tag)

	assert.Equal(t, "name", column.Name)
	assert.Equal(t, "varchar(10)", column.Type)
	assert.Equal(t, true, column.Nullable)
}

func TestUnMarshallColumnTag_IdFromSequence(t *testing.T) {
	tag := "name:id;type:integer;primaryKey;autoIncrement;nullable:false;default:nextval('activity_logs_id_seq')"

	column := raiden.UnmarshalColumnTag(tag)

	assert.Equal(t, "id", column.Name)
	assert.Equal(t, "integer", column.Type)
	assert.Equal(t, false, column.Nullable)
}

func TestUnMarshallJoinTag(t *testing.T) {
	hasOneTag := `joinType:hasOne;primaryKey:id;foreignKey:scouter_id`
	hasOne := raiden.UnmarshalJoinTag(hasOneTag)
	assert.Equal(t, raiden.RelationTypeHasOne, hasOne.JoinType)
	assert.Equal(t, "id", hasOne.PrimaryKey)
	assert.Equal(t, "scouter_id", hasOne.ForeignKey)

	hasManyTag := `joinType:hasMany;primaryKey:id;foreignKey:scouter_id`
	hasMany := raiden.UnmarshalJoinTag(hasManyTag)
	assert.Equal(t, raiden.RelationTypeHasMany, hasMany.JoinType)
	assert.Equal(t, "id", hasMany.PrimaryKey)
	assert.Equal(t, "scouter_id", hasMany.ForeignKey)

	manyToManyTag := `joinType:manyToMany;through:submission;sourcePrimaryKey:id;sourceForeignKey:candidate_id;targetPrimaryKey:id;targetForeign:candidate_id`
	manyToMany := raiden.UnmarshalJoinTag(manyToManyTag)
	assert.Equal(t, raiden.RelationTypeManyToMany, manyToMany.JoinType)
	assert.Equal(t, "id", manyToMany.SourcePrimaryKey)
	assert.Equal(t, "candidate_id", manyToMany.SourceForeignKey)
	assert.Equal(t, "id", manyToMany.TargetPrimaryKey)
	assert.Equal(t, "candidate_id", manyToMany.TargetForeignKey)
	assert.Equal(t, "submission", manyToMany.Through)
}

// ---- Test get table name
type Event struct {
	raiden.ModelBase
}

type EventSource struct {
	raiden.ModelBase
}

type EventCalendar struct {
	raiden.ModelBase
	Metadata string `tableName:"EventCalendar"`
}

func TestGetTableName_ReturnDefault(t *testing.T) {
	tableName := raiden.GetTableName(Event{})
	assert.Equal(t, "event", tableName)
}

func TestGetTableName_ReturnMustBeSnakeCase(t *testing.T) {
	tableName := raiden.GetTableName(EventSource{})
	assert.Equal(t, "event_source", tableName)
}

func TestGetTableName_ReturnNameFromTag(t *testing.T) {
	tableName := raiden.GetTableName(EventCalendar{})
	assert.Equal(t, "EventCalendar", tableName)
}
