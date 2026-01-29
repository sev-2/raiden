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
}
