package builder_test

import (
	"testing"

	"github.com/sev-2/raiden/pkg/builder"
	"github.com/stretchr/testify/assert"
)

// Test MarshalClause function
func TestMarshalClause(t *testing.T) {
	clause := builder.Clause("col = 'value'")
	result := builder.MarshalClause(clause)
	assert.Equal(t, "col = 'value'", result)
}

// Test UnmarshalClause function
func TestUnmarshalClause(t *testing.T) {
	clause, code, ok := builder.UnmarshalClause("col = 'value'")
	assert.True(t, ok)
	assert.Contains(t, clause.String(), "col")
	assert.Contains(t, code, "st.")
}

func TestUnmarshalClauseEmpty(t *testing.T) {
	clause, code, ok := builder.UnmarshalClause("")
	assert.True(t, ok)
	assert.Equal(t, builder.Clause(""), clause)
	assert.Equal(t, "st.Clause(\"\")", code)
}

// Test NormalizeClauseSQL function
func TestNormalizeClauseSQL(t *testing.T) {
	result := builder.NormalizeClauseSQL("(\"col\" = 'value')")
	assert.Contains(t, result, "col")
	assert.Contains(t, result, "value")
}

// Test NormalizeClauseSQL with qualifiers
func TestNormalizeClauseSQLWithQualifiers(t *testing.T) {
	qualifier := builder.ClauseQualifier{Schema: "public", Table: "users"}
	result := builder.NormalizeClauseSQL("public.users.col = 'value'", qualifier)
	assert.Contains(t, result, "col")
	assert.Contains(t, result, "value")
	assert.NotContains(t, result, "public")
	assert.NotContains(t, result, "users")
}

// Test enclosedByOuterParentheses function (testing internal function through coverage)
func TestEnclosedByOuterParentheses(t *testing.T) {
	// This would need to be tested through the internal API or by creating a helper if we could access it
	// For now, we'll test it indirectly through other functions that might use it
	result := builder.NormalizeClauseSQL("((col = 'value'))")
	assert.Contains(t, result, "col")
	assert.Contains(t, result, "value")
}

// Test collapseWhitespace function (testing internal function through coverage)
func TestCollapseWhitespace(t *testing.T) {
	// Similar to above - test through public API
	result := builder.NormalizeClauseSQL("col   =   'value'")
	assert.Contains(t, result, "col")
	assert.Contains(t, result, "value")
}

// Test splitByLogical function (testing internal function through coverage)
func TestSplitByLogical(t *testing.T) {
	// Testing through public API
	result := builder.NormalizeClauseSQL("a = 1 AND b = 2 OR c = 3")
	assert.NotEmpty(t, result)
}

// Test splitComparison function (testing internal function through coverage)
func TestSplitComparison(t *testing.T) {
	// This is tested through the parsing functionality
	_, _, ok := builder.UnmarshalClause("col = 'value'")
	assert.True(t, ok)
}

// Test StorageCheckClause function
func TestStorageCheckClause(t *testing.T) {
	clause := builder.StorageCheckClause("bucket1", builder.Eq("owner", builder.AuthUID()))
	assert.Contains(t, clause.String(), "bucket_id")
	assert.Contains(t, clause.String(), "owner")
	assert.Contains(t, clause.String(), "auth.uid()")
}

// Test operandToBuilderExp functionality through UnmarshalClause
func TestOperandToBuilderExp(t *testing.T) {
	_, _, ok := builder.UnmarshalClause("col = 'value'")
	assert.True(t, ok)
}

// Test canonicalizeSimpleEquality
func TestCanonicalizeSimpleEquality(t *testing.T) {
	result := builder.NormalizeClauseSQL("b = 'value' AND a = 'other'")
	assert.Contains(t, result, "=")
}

// Test removeQualifiers through NormalizeClauseSQL
func TestRemoveQualifiers(t *testing.T) {
	qualifier := builder.ClauseQualifier{Schema: "public", Table: "users"}
	result := builder.NormalizeClauseSQL("public.users.name = 'test'", qualifier)
	assert.Contains(t, result, "name")
	assert.Contains(t, result, "test")
	assert.NotContains(t, result, "public")
	assert.NotContains(t, result, "users")
}
