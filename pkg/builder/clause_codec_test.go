package builder_test

import (
	"testing"

	b "github.com/sev-2/raiden/pkg/builder"
)

func TestNormalizeClauseSQL_StripsFormattingNoise(t *testing.T) {
	qualifiers := []b.ClauseQualifier{{Schema: "public", Table: "table_normalized"}}

	lhs := b.NormalizeClauseSQL("auth.uid = owner_id", qualifiers...)
	rhs := b.NormalizeClauseSQL(`((("public"."table_normalized"."owner_id")::text) = (auth.uid)::text)`, qualifiers...)

	if lhs != rhs {
		t.Fatalf("expected normalized clauses to match, got %q vs %q", lhs, rhs)
	}
}

func TestNormalizeClauseSQL_Optional(t *testing.T) {
	qualifiers := []b.ClauseQualifier{{Table: "policies"}}

	clause := `(("policies"."role") = 'anon')`
	normalized := b.NormalizeClauseSQL(clause, qualifiers...)
	if normalized != "'anon' = role" {
		t.Fatalf("unexpected normalized clause: %q", normalized)
	}
}

func TestUnmarshalClause_SimpleEquality(t *testing.T) {
	qualifiers := []b.ClauseQualifier{{Schema: "public", Table: "courses"}}
	clause, code, ok := b.UnmarshalClause("auth.uid() = public.courses.owner_id", qualifiers...)
	if !ok {
		t.Fatalf("expected builder conversion to succeed")
	}
	if clause.String() != b.Eq("owner_id", b.AuthUID()).String() {
		t.Fatalf("unexpected clause value: %q", clause)
	}
	if code != `st.Eq("owner_id", st.AuthUID())` {
		t.Fatalf("unexpected builder code: %q", code)
	}
}

func TestUnmarshalClause_StringLiteral(t *testing.T) {
	qualifiers := []b.ClauseQualifier{{Table: "policies"}}
	clause, code, ok := b.UnmarshalClause(`role = 'anon'`, qualifiers...)
	if !ok {
		t.Fatalf("expected builder conversion to succeed")
	}
	expected := b.Eq("role", b.String("anon"))
	if clause.String() != expected.String() {
		t.Fatalf("unexpected clause: %q", clause)
	}
	if code != `st.Eq("role", st.String("anon"))` {
		t.Fatalf("unexpected builder code: %q", code)
	}
}

func TestUnmarshalClause_AndCombination(t *testing.T) {
	qualifiers := []b.ClauseQualifier{{Schema: "public", Table: "courses"}}
	expr := "auth.uid() = owner_id AND role = 'admin'"
	clause, code, ok := b.UnmarshalClause(expr, qualifiers...)
	if !ok {
		t.Fatalf("expected builder conversion to succeed")
	}
	expected := b.And(b.Eq("owner_id", b.AuthUID()), b.Eq("role", b.String("admin")))
	if clause.String() != expected.String() {
		t.Fatalf("unexpected clause: %q", clause)
	}
	if code != `st.And(st.Eq("owner_id", st.AuthUID()), st.Eq("role", st.String("admin")))` {
		t.Fatalf("unexpected builder code: %q", code)
	}
}

func TestUnmarshalClause_Fallback(t *testing.T) {
	qualifiers := []b.ClauseQualifier{{Schema: "public", Table: "courses"}}
	expr := "(owner_id = auth.uid()) OR EXISTS(SELECT 1)"
	_, _, ok := b.UnmarshalClause(expr, qualifiers...)
	if ok {
		t.Fatalf("expected conversion to fail for complex expression")
	}
}
