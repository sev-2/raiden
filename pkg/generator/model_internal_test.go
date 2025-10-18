package generator

import (
	"strings"
	"testing"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/builder"
	"github.com/sev-2/raiden/pkg/postgres"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/require"
)

func TestMapTableAttributes_AddsImportsAndTags(t *testing.T) {
	table := objects.Table{
		Name: "sample_table",
		Columns: []objects.Column{
			{Name: "id", DataType: string(postgres.IntType), IsNullable: false},
			{Name: "created_at", DataType: string(postgres.TimestampTzType), IsNullable: true},
			{Name: "status", DataType: string(postgres.UserDefined), Format: "status_type", IsNullable: true},
		},
		PrimaryKeys: []objects.PrimaryKey{{Name: "id"}},
	}

	mapTypes := map[string]objects.Type{
		"status_type": {Name: "status_type", Format: "status_type"},
	}

	validation := state.ModelValidationTag{"id": "required"}

	cols, imports := MapTableAttributes("github.com/example/project", table, mapTypes, validation)

	require.Len(t, cols, 3)
	require.Equal(t, "id", cols[0].Name)
	require.Contains(t, cols[0].Tag, "primaryKey")
	require.Contains(t, cols[0].Tag, "validate:\"required\"")

	require.Equal(t, "*postgres.DateTime", cols[1].Type)
	require.Equal(t, "types.StatusType", cols[2].Type)

	require.Contains(t, imports, "github.com/sev-2/raiden/pkg/postgres")
	require.NotContains(t, imports, "github.com/google/uuid")
	require.Contains(t, imports, "project/internal/types")
}

func TestGenerateClauseCodeHandlesVariants(t *testing.T) {
	tableMap := map[string]objects.Table{
		"public.sample": {
			Columns: []objects.Column{{Name: "owner_id"}},
		},
	}
	qualifier := builder.ClauseQualifier{Schema: "public", Table: "sample"}

	code := generateClauseCode("TRUE", qualifier, "m", nil)
	require.Equal(t, "st.True", code)

	sql := "\"owner_id\" = 'abc'"
	code = generateClauseCode(sql, qualifier, "model", tableMap["public.sample"].Columns)
	require.Contains(t, code, "st.ColOf(model, model.OwnerId)")

	fallback := generateClauseCode("invalid syntax $", qualifier, "", nil)
	require.Equal(t, `st.Clause("invalid syntax $")`, fallback)
}

func TestBuildColumnAndRelationHelpers(t *testing.T) {
	col := objects.Column{
		Name:         "is_active",
		DataType:     string(postgres.BooleanType),
		IsNullable:   false,
		IsUnique:     true,
		DefaultValue: "true",
	}
	tag := buildColumnTag(col, map[string]bool{"is_active": true}, nil, nil)
	require.Contains(t, tag, "primaryKey")
	require.Contains(t, tag, "default:true")
	require.Contains(t, tag, "unique")

	rel := state.Relation{
		Table:        "profile",
		RelationType: raiden.RelationTypeHasMany,
		ForeignKey:   "profile_id",
		PrimaryKey:   "id",
		Action:       &objects.TablesRelationshipAction{UpdateAction: "c", DeletionAction: "r"},
	}
	relTag := BuildRelationTag(&rel)
	require.Contains(t, relTag, "joinType:")
	require.Contains(t, relTag, "primaryKey:id")
	require.Contains(t, relTag, "onUpdate:\"cascade\"")

	fields := BuildRelationFields(objects.Table{Name: "profile"}, []state.Relation{rel})
	require.Equal(t, 1, len(fields))
	require.True(t, strings.Contains(fields[0].Table, "Profiles"))
}
