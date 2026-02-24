package state_test

import (
	"testing"
	"time"

	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
)

// CandidateAlias is a second struct pointing to the same table as Candidate.
// Used to test duplicate table name registration handling.
type CandidateAlias struct {
	Id   int64   `json:"id,omitempty" column:"name:id;type:bigint;primaryKey;autoIncrement;nullable:false"`
	Name *string `json:"name,omitempty" column:"name:name;type:varchar;nullable;unique"`

	// Table information — intentionally same tableName as Candidate
	Metadata string `json:"-" schema:"public" tableName:"candidate"`

	// Access control
	Acl string `json:"-" read:"anon" write:"anon"`
}

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
	Acl string `json:"-" read:"anon" write:"anon"`

	// Relations
	Candidate *Candidate `json:"candidate,omitempty" join:"joinType:hasOne;primaryKey:id;foreignKey:candidate_id"  onUpdate:"cascade" onDelete:"cascade"`
}

type Candidate struct {
	Id        int64      `json:"id,omitempty" column:"name:id;type:bigint;primaryKey;autoIncrement;nullable:false"`
	Name      *string    `json:"name,omitempty" column:"name:name;type:varchar;nullable;unique"`
	Batch     *int32     `json:"batch,omitempty" column:"name:batch;type:integer;nullable"`
	CreatedAt *time.Time `json:"created_at,omitempty" column:"name:created_at;type:timestampz;nullable;default:now()"`

	// Table information
	Metadata string `json:"-" schema:"public" replicaIdentity:"DEFAULT"`

	// Access control
	Acl string `json:"-" read:"anon" write:"authenticated"`

	// Relations
	Submission []*Submission `json:"submission,omitempty" join:"joinType:hasMany;primaryKey:id;foreignKey:candidate_id"  onUpdate:"cascade" onDelete:"cascade"`
}

func TestExtractTable_NoRelation(t *testing.T) {
	tableState := make([]state.TableState, 0)
	appTable := []any{&Candidate{}}
	rs, err := state.ExtractTable(tableState, appTable, nil)

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
	rs, err := state.ExtractTable(tableState, appTable, nil)

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

func TestExtractTable(t *testing.T) {
	relationAction := objects.TablesRelationshipAction{
		UpdateAction:   "cascade",
		DeletionAction: "cascade",
	}

	tableStates := []state.TableState{
		{
			Table: objects.Table{
				Name: "submission",
				Relationships: []objects.TablesRelationship{
					{
						SourceSchema:      "public",
						SourceTableName:   "submission",
						SourceColumnName:  "candidate_id",
						TargetTableSchema: "public",
						TargetTableName:   "candidate",
						TargetColumnName:  "id",
						Action:            &relationAction,
					},
				},
				PrimaryKeys: []objects.PrimaryKey{
					{
						Name:      "id",
						Schema:    "public",
						TableName: "submission",
					},
				},
				Columns: []objects.Column{
					{
						Name: "id",
					},
					{
						Name: "scouter_id",
					},
					{
						Name: "candidate_id",
					},
					{
						Name: "score",
					},
				},
			},
		},
	}

	appTable := []any{&Submission{}}
	result, err := state.ExtractTable(tableStates, appTable, nil)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(result.New))
	assert.Equal(t, 1, len(result.Existing))
	assert.Equal(t, 0, len(result.Delete))
}

func TestToFlatTable(t *testing.T) {
	items := state.ExtractTableItems{
		{
			Table: objects.Table{Name: "table1"},
		},
		{
			Table: objects.Table{Name: "table2"},
		},
	}

	tables := items.ToFlatTable()
	assert.Equal(t, 2, len(tables))
	assert.Equal(t, "table1", tables[0].Name)
	assert.Equal(t, "table2", tables[1].Name)
}

func TestToDeleteFlatMap(t *testing.T) {
	result := state.ExtractTableResult{
		Delete: state.ExtractTableItems{
			{
				Table: objects.Table{Name: "table1"},
			},
			{
				Table: objects.Table{Name: "table2"},
			},
		},
	}

	mapData := result.ToDeleteFlatMap()
	assert.Equal(t, 2, len(mapData))
	assert.Equal(t, "table1", mapData["table1"].Name)
	assert.Equal(t, "table2", mapData["table2"].Name)
}

func TestExtractTable_DuplicateTableName(t *testing.T) {
	// Two models (Candidate + CandidateAlias) point to the same tableName "candidate".
	// The second registration should be skipped, not treated as "New".
	tableState := []state.TableState{
		{
			Table: objects.Table{
				ID:   100,
				Name: "candidate",
				PrimaryKeys: []objects.PrimaryKey{
					{Name: "id", Schema: "public", TableName: "candidate"},
				},
				Columns: []objects.Column{{Name: "id"}, {Name: "name"}},
			},
		},
	}
	appTable := []any{&Candidate{}, &CandidateAlias{}}
	rs, err := state.ExtractTable(tableState, appTable, nil)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(rs.Existing), "should have exactly one Existing entry")
	assert.Equal(t, 0, len(rs.New), "duplicate should NOT create a New entry")
	assert.Equal(t, 0, len(rs.Delete))
	assert.Equal(t, 100, rs.Existing[0].Table.ID)
}

func TestExtractTable_DuplicateTableName_NoState(t *testing.T) {
	// Two models with same tableName, neither in state — only one New entry expected.
	appTable := []any{&Candidate{}, &CandidateAlias{}}
	rs, err := state.ExtractTable(nil, appTable, nil)

	assert.NoError(t, err)
	assert.Equal(t, 0, len(rs.Existing))
	assert.Equal(t, 1, len(rs.New), "duplicate should NOT create a second New entry")
	assert.Equal(t, "candidate", rs.New[0].Table.Name)
}
