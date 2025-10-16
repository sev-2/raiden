package utils_test

import (
	"testing"

	"github.com/sev-2/raiden/pkg/builder"
	"github.com/sev-2/raiden/pkg/utils"
)

func TestNormalizeExpression_StripsFormattingNoise(t *testing.T) {
	qualifiers := []utils.ExpressionQualifier{
		{Schema: "public", Table: "table_normalized"},
	}

	lhs := utils.NormalizeExpression("auth.uid = owner_id", qualifiers...)
	rhs := utils.NormalizeExpression(`((("public"."table_normalized"."owner_id")::text) = (auth.uid)::text)`, qualifiers...)

	if lhs != rhs {
		t.Fatalf("expected normalized clauses to match, got %q vs %q", lhs, rhs)
	}
}

func TestNormalizeOptionalExpression(t *testing.T) {
	qualifiers := []utils.ExpressionQualifier{{Table: "policies"}}

	clause := `(("policies"."role") = 'anon')`
	normalized, ok := utils.NormalizeOptionalExpression(&clause, qualifiers...)
	if !ok {
		t.Fatalf("expected clause to be present")
	}
	if normalized != "'anon' = role" {
		t.Fatalf("unexpected normalized clause: %q", normalized)
	}

	if _, ok := utils.NormalizeOptionalExpression(nil, qualifiers...); ok {
		t.Fatalf("expected empty clause to return !ok")
	}
}

func TestNormalizeExpression_CanonicalEquality(t *testing.T) {
	clauseA := utils.NormalizeExpression("owner_id = auth.uid")
	clauseB := utils.NormalizeExpression("auth.uid = owner_id")

	if clauseA != clauseB {
		t.Fatalf("expected canonical equality, got %q vs %q", clauseA, clauseB)
	}
}

func TestNormalizeExpression_WithBuilderOutput(t *testing.T) {
	qualifier := utils.ExpressionQualifier{Schema: "public", Table: "courses"}

	built := builder.Eq("public.courses.owner_id", builder.AuthUID())
	normalized := utils.NormalizeExpression(built.String(), qualifier)

	expected := utils.NormalizeExpression("auth.uid() = public.courses.owner_id", qualifier)
	if normalized != expected {
		t.Fatalf("builder clause normalization mismatch, got %q expected %q", normalized, expected)
	}
}
