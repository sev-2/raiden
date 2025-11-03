package query

import (
	"testing"

	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
)

func TestBuildFunctionQueryCreate(t *testing.T) {
	fn := &objects.Function{
		Schema:            "public",
		Name:              "hello",
		CompleteStatement: "CREATE OR REPLACE FUNCTION public.hello() RETURNS void LANGUAGE plpgsql AS $$ BEGIN END; $$",
	}

	stmt, err := BuildFunctionQuery(FunctionActionCreate, fn)
	assert.NoError(t, err)
	assert.Contains(t, stmt, "CREATE OR REPLACE FUNCTION public.hello()")
	assert.True(t, stmt[len(stmt)-1] == ';')
}

func TestBuildFunctionQueryDelete(t *testing.T) {
	fn := &objects.Function{Schema: "public", Name: "hello", IdentityArgumentTypes: "integer"}
	stmt, err := BuildFunctionQuery(FunctionActionDelete, fn)
	assert.NoError(t, err)
	assert.Equal(t, "DROP FUNCTION IF EXISTS \"public\".\"hello\"(integer);", stmt)
}

func TestBuildFunctionQueryUpdate(t *testing.T) {
	fn := &objects.Function{
		Schema:                "public",
		Name:                  "hello",
		IdentityArgumentTypes: "",
		CompleteStatement:     "CREATE OR REPLACE FUNCTION public.hello() RETURNS void LANGUAGE sql AS $$ SELECT 1; $$",
	}

	stmt, err := BuildFunctionQuery(FunctionActionUpdate, fn)
	assert.NoError(t, err)
	assert.Contains(t, stmt, "DROP FUNCTION IF EXISTS \"public\".\"hello\"();")
	assert.Contains(t, stmt, "CREATE OR REPLACE FUNCTION public.hello()")
}

func TestBuildFunctionQueryValidation(t *testing.T) {
	_, err := BuildFunctionQuery(FunctionActionCreate, nil)
	assert.Error(t, err)

	_, err = BuildFunctionQuery(FunctionActionCreate, &objects.Function{Schema: "public", Name: "x"})
	assert.Error(t, err)
}
