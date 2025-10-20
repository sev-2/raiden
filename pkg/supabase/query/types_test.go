package query

import (
	"testing"

	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
)

func TestBuildCreateTypeQuery(t *testing.T) {
	stmt := BuildCreateTypeQuery(&objects.Type{Schema: "public", Name: "status", Enums: []string{"draft", "published"}})
	assert.Contains(t, stmt, "CREATE TYPE IF NOT EXISTS \"public\".\"status\" AS ENUM ('draft','published')")
}

func TestBuildDeleteTypeQuery(t *testing.T) {
	stmt := BuildDeleteTypeQuery(&objects.Type{Schema: "public", Name: "status"})
	assert.Equal(t, "DROP TYPE IF EXISTS \"public\".\"status\" CASCADE;", stmt)
}

func TestBuildTypeQueryUpdate(t *testing.T) {
	typ := &objects.Type{Schema: "public", Name: "status", Enums: []string{"draft"}}
	stmt, err := BuildTypeQuery(TypeActionUpdate, typ)
	assert.NoError(t, err)
	assert.Contains(t, stmt, "DROP TYPE IF EXISTS \"public\".\"status\" CASCADE;")
	assert.Contains(t, stmt, "CREATE TYPE IF NOT EXISTS \"public\".\"status\"")
}

func TestBuildTypeQueryValidation(t *testing.T) {
	_, err := BuildTypeQuery("invalid", &objects.Type{})
	assert.Error(t, err)
}

func TestBuildTypeQueryCreateAndDelete(t *testing.T) {
	typ := &objects.Type{Schema: "public", Name: "status", Enums: []string{"draft"}}

	createStmt, err := BuildTypeQuery(TypeActionCreate, typ)
	assert.NoError(t, err)
	assert.Equal(t, BuildCreateTypeQuery(typ), createStmt)

	deleteStmt, err := BuildTypeQuery(TypeActionDelete, typ)
	assert.NoError(t, err)
	assert.Equal(t, BuildDeleteTypeQuery(typ), deleteStmt)
}
