package state_test

import (
	"testing"
	"time"

	"github.com/sev-2/raiden/pkg/state"
	"github.com/stretchr/testify/assert"
)

type Submission struct {
	Id          int64      `json:"id,omitempty" column:"name:id;type:bigint;primaryKey;autoIncrement;nullable:false"`
	ScouterId   *int64     `json:"scouter_id,omitempty" column:"name:scouter_id;type:bigint;nullable"`
	CandidateId *int64     `json:"candidate_id,omitempty" column:"name:candidate_id;type:bigint;nullable"`
	Score       *float64   `json:"score,omitempty" column:"name:score;type:real;nullable"`
	Note        *string    `json:"note,omitempty" column:"name:note;type:text;nullable"`
	CreatedAt   *time.Time `json:"created_at,omitempty" column:"name:created_at;type:timestampz;nullable;default:now()"`

	// Table information
	Metadata string `json:"-" schema:"public"`

	// Access control
	Acl string `json:"-" read:"" write:""`

	// Relations
	Candidate *Candidate `json:"candidate,omitempty" join:"joinType:hasOne;primaryKey:id;foreignKey:candidate_id"`
}

type Candidate struct {
	Id        int64      `json:"id,omitempty" column:"name:id;type:bigint;primaryKey;autoIncrement;nullable:false"`
	Name      *string    `json:"name,omitempty" column:"name:name;type:varchar;nullable;unique"`
	Batch     *int32     `json:"batch,omitempty" column:"name:batch;type:integer;nullable"`
	CreatedAt *time.Time `json:"created_at,omitempty" column:"name:created_at;type:timestampz;nullable;default:now()"`

	// Table information
	Metadata string `json:"-" schema:"public" replicaIdentity:"DEFAULT"`

	// Access control
	Acl string `json:"-" read:"" write:""`

	// Relations
	Submission []*Submission `json:"submission,omitempty" join:"joinType:hasMany;primaryKey:id;foreignKey:candidate_id"`
}

func TestExtractTable_NoRelation(t *testing.T) {
	tableState := make([]state.TableState, 0)
	appTable := []any{&Candidate{}}
	rs, err := state.ExtractTable(tableState, appTable)

	assert.NoError(t, err)
	assert.Equal(t, 0, len(rs.Existing))
	assert.Equal(t, 0, len(rs.Delete))
	assert.Equal(t, 1, len(rs.New))

	// assert table
	assert.Equal(t, "candidate", rs.New[0].Table.Name)
	assert.Equal(t, "public", rs.New[0].Table.Schema)

	// assert pk
	assert.Equal(t, 1, len(rs.New[0].Table.PrimaryKeys))
	assert.Equal(t, "id", rs.New[0].Table.PrimaryKeys[0].Name)
	assert.Equal(t, "public", rs.New[0].Table.PrimaryKeys[0].Schema)
	assert.Equal(t, "candidate", rs.New[0].Table.PrimaryKeys[0].TableName)

	// assert column
	assert.Equal(t, 4, len(rs.New[0].Table.Columns))
	assert.Equal(t, "id", rs.New[0].Table.Columns[0].Name)
	assert.Equal(t, "name", rs.New[0].Table.Columns[1].Name)
	assert.Equal(t, "batch", rs.New[0].Table.Columns[2].Name)
	assert.Equal(t, "created_at", rs.New[0].Table.Columns[3].Name)
}

func TestExtractTable_WithRelation(t *testing.T) {
	tableState := make([]state.TableState, 0)
	appTable := []any{&Submission{}}
	rs, err := state.ExtractTable(tableState, appTable)

	assert.NoError(t, err)
	assert.Equal(t, 0, len(rs.Existing))
	assert.Equal(t, 0, len(rs.Delete))
	assert.Equal(t, 1, len(rs.New))

	// assert table
	assert.Equal(t, "submission", rs.New[0].Table.Name)
	assert.Equal(t, "public", rs.New[0].Table.Schema)

	// assert pk
	assert.Equal(t, 1, len(rs.New[0].Table.PrimaryKeys))
	assert.Equal(t, "id", rs.New[0].Table.PrimaryKeys[0].Name)
	assert.Equal(t, "public", rs.New[0].Table.PrimaryKeys[0].Schema)
	assert.Equal(t, "submission", rs.New[0].Table.PrimaryKeys[0].TableName)

	// assert relation
	assert.Equal(t, 1, len(rs.New[0].Table.Relationships))
	assert.Equal(t, "public", rs.New[0].Table.Relationships[0].SourceSchema)
	assert.Equal(t, "submission", rs.New[0].Table.Relationships[0].SourceTableName)
	assert.Equal(t, "candidate_id", rs.New[0].Table.Relationships[0].SourceColumnName)
	assert.Equal(t, "public", rs.New[0].Table.Relationships[0].TargetTableSchema)
	assert.Equal(t, "candidate", rs.New[0].Table.Relationships[0].TargetTableName)
	assert.Equal(t, "id", rs.New[0].Table.Relationships[0].TargetColumnName)

	// assert column
	assert.Equal(t, 6, len(rs.New[0].Table.Columns))
	assert.Equal(t, "id", rs.New[0].Table.Columns[0].Name)
	assert.Equal(t, "scouter_id", rs.New[0].Table.Columns[1].Name)
	assert.Equal(t, "candidate_id", rs.New[0].Table.Columns[2].Name)
	assert.Equal(t, "score", rs.New[0].Table.Columns[3].Name)
	assert.Equal(t, "note", rs.New[0].Table.Columns[4].Name)
	assert.Equal(t, "created_at", rs.New[0].Table.Columns[5].Name)
}
