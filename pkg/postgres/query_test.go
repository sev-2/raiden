package postgres_test

import (
	"testing"

	"github.com/sev-2/raiden/pkg/postgres"
	"github.com/stretchr/testify/assert"
)

func TestIsReservedKeyword(t *testing.T) {
	// Test for reserved keywords
	reservedKeywords := []string{
		postgres.Select, postgres.From, postgres.Where, postgres.OrderBy, postgres.Limit, postgres.Offset,
		postgres.GroupBy, postgres.Having, postgres.Insert, postgres.Update, postgres.Delete, postgres.Create,
		postgres.Alter, postgres.Drop, postgres.Truncate, postgres.Join, postgres.Inner, postgres.Outer,
		postgres.Left, postgres.Right, postgres.LeftJoin, postgres.RightJoin, postgres.InnerJoin, postgres.OuterJoin,
		postgres.On, postgres.As, postgres.And, postgres.Or, postgres.Not, postgres.Between, postgres.In, postgres.Like,
		postgres.Exists, postgres.All, postgres.Any, postgres.Union, postgres.Intersect, postgres.Except,
		postgres.Asc, postgres.Desc, postgres.Is, postgres.IsNull, postgres.IsNotNull,
		postgres.Case, postgres.When, postgres.Then, postgres.Else, postgres.End, postgres.With,
	}

	for _, keyword := range reservedKeywords {
		t.Run(keyword, func(t *testing.T) {
			assert.True(t, postgres.IsReservedKeyword(keyword))
		})
	}

	// Test for non-reserved keywords
	nonReservedKeywords := []string{
		"RANDOMKEYWORD1", "randomKeyword2", "anotherkeyword", "nonreserved",
	}

	for _, keyword := range nonReservedKeywords {
		t.Run(keyword, func(t *testing.T) {
			assert.False(t, postgres.IsReservedKeyword(keyword))
		})
	}
}

func TestIsReservedSymbol(t *testing.T) {
	// Test for reserved symbols
	reservedSymbols := []string{
		"=", "<>", "!=", ">", "<", ">=", "<=", "+", "-", "*", "/", "(", ")", ",", ";", ".", ":", "::", "::=", "||",
		"<=>", "&", "|", "^", "<<", ">>", "~", "<<=", ">>=", "&=", "|=", "^=", "~=", "%", "@", "#", "$", "`", "[", "]",
		"{", "}", "!", "?", ":=", "=>",
	}

	for _, symbol := range reservedSymbols {
		t.Run(symbol, func(t *testing.T) {
			assert.True(t, postgres.IsReservedSymbol(symbol))
		})
	}

	// Test for non-reserved symbols
	nonReservedSymbols := []string{
		"abc", "123", "table_name", "WORD",
	}

	for _, symbol := range nonReservedSymbols {
		t.Run(symbol, func(t *testing.T) {
			assert.False(t, postgres.IsReservedSymbol(symbol))
		})
	}
}

func TestIsReservedKeyword_CaseInsensitivity(t *testing.T) {
	// Test that keywords are case-sensitive (uppercase)
	testCases := []struct {
		keyword  string
		expected bool
		desc     string
	}{
		{"SELECT", true, "uppercase SELECT should be reserved"},
		{"select", false, "lowercase select should not be reserved"},
		{"IS", true, "uppercase IS should be reserved"},
		{"is", false, "lowercase is should not be reserved"},
		{"CASE", true, "uppercase CASE should be reserved"},
		{"case", false, "lowercase case should not be reserved"},
		{"WHEN", true, "uppercase WHEN should be reserved"},
		{"when", false, "lowercase when should not be reserved"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			result := postgres.IsReservedKeyword(tc.keyword)
			assert.Equal(t, tc.expected, result, tc.desc)
		})
	}
}

func TestNewKeywords_ISKeyword(t *testing.T) {
	// Test IS keyword specifically
	t.Run("IS is reserved", func(t *testing.T) {
		assert.True(t, postgres.IsReservedKeyword(postgres.Is))
	})

	t.Run("IS NULL is reserved", func(t *testing.T) {
		assert.True(t, postgres.IsReservedKeyword(postgres.IsNull))
	})

	t.Run("IS NOT NULL is reserved", func(t *testing.T) {
		assert.True(t, postgres.IsReservedKeyword(postgres.IsNotNull))
	})

	// Verify the constant values
	t.Run("IS constant value", func(t *testing.T) {
		assert.Equal(t, "IS", postgres.Is)
	})
}

func TestNewKeywords_CaseExpressionKeywords(t *testing.T) {
	// Test CASE expression keywords
	caseKeywords := []struct {
		keyword  string
		expected string
	}{
		{postgres.Case, "CASE"},
		{postgres.When, "WHEN"},
		{postgres.Then, "THEN"},
		{postgres.Else, "ELSE"},
		{postgres.End, "END"},
	}

	for _, tc := range caseKeywords {
		t.Run(tc.keyword+" is reserved", func(t *testing.T) {
			assert.True(t, postgres.IsReservedKeyword(tc.keyword))
			assert.Equal(t, tc.expected, tc.keyword)
		})
	}
}

func TestReservedKeywords_AllDeclaredConstantsAreReserved(t *testing.T) {
	// Ensure all declared constants are in the reserved map
	declaredKeywords := []string{
		postgres.Select, postgres.From, postgres.Where, postgres.OrderBy, postgres.Limit, postgres.Offset,
		postgres.GroupBy, postgres.Having, postgres.Insert, postgres.Update, postgres.Delete, postgres.Create,
		postgres.Alter, postgres.Drop, postgres.Truncate, postgres.Join, postgres.Inner, postgres.Outer,
		postgres.Left, postgres.Right, postgres.LeftJoin, postgres.RightJoin, postgres.InnerJoin, postgres.OuterJoin,
		postgres.On, postgres.As, postgres.And, postgres.Or, postgres.Not, postgres.Between, postgres.In, postgres.Like,
		postgres.Exists, postgres.All, postgres.Any, postgres.Union, postgres.Intersect, postgres.Except,
		postgres.Asc, postgres.Desc, postgres.Is, postgres.IsNull, postgres.IsNotNull,
		postgres.Case, postgres.When, postgres.Then, postgres.Else, postgres.End, postgres.With,
	}

	for _, keyword := range declaredKeywords {
		t.Run("Declared_"+keyword, func(t *testing.T) {
			assert.True(t, postgres.IsReservedKeyword(keyword),
				"Constant %s should be in ReservedKeywords map", keyword)
		})
	}
}

func TestReservedKeywords_JoinVariants(t *testing.T) {
	// Test all JOIN variant keywords
	joinKeywords := map[string]string{
		"base JOIN":  postgres.Join,
		"INNER":      postgres.Inner,
		"OUTER":      postgres.Outer,
		"LEFT":       postgres.Left,
		"RIGHT":      postgres.Right,
		"LEFT JOIN":  postgres.LeftJoin,
		"RIGHT JOIN": postgres.RightJoin,
		"INNER JOIN": postgres.InnerJoin,
		"OUTER JOIN": postgres.OuterJoin,
	}

	for desc, keyword := range joinKeywords {
		t.Run(desc, func(t *testing.T) {
			assert.True(t, postgres.IsReservedKeyword(keyword))
		})
	}
}

func TestReservedKeywords_SQLiteralUsage(t *testing.T) {
	// Test that keywords appear correctly in SQL-like contexts
	tests := []struct {
		name    string
		keyword string
		value   string
	}{
		{"IS in condition", postgres.Is, "IS"},
		{"CASE expression", postgres.Case, "CASE"},
		{"WHEN clause", postgres.When, "WHEN"},
		{"THEN clause", postgres.Then, "THEN"},
		{"ELSE clause", postgres.Else, "ELSE"},
		{"END expression", postgres.End, "END"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.value, tt.keyword)
			assert.True(t, postgres.IsReservedKeyword(tt.keyword))
		})
	}
}
