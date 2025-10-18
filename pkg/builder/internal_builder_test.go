package builder

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFieldMatchScoreVariants(t *testing.T) {
	type embedded struct{}
	type sample struct {
		Column   string `column:"name:custom" json:"custom_value"`
		JsonOnly string `json:"json_only"`
		embedded
	}

	typ := reflect.TypeOf(sample{})

	sfColumn, ok := typ.FieldByName("Column")
	require.True(t, ok)
	require.Equal(t, 260, fieldMatchScore(sfColumn))

	sfJSON, ok := typ.FieldByName("JsonOnly")
	require.True(t, ok)
	require.Equal(t, 60, fieldMatchScore(sfJSON))

	sfEmbedded, ok := typ.FieldByName("embedded")
	require.True(t, ok)
	require.Equal(t, 0, fieldMatchScore(sfEmbedded))
}

func TestColumnHelperFunctions(t *testing.T) {
	type sample struct {
		ColumnTag string `column:"name:explicit"`
		JSONTag   string `json:"json_name,omitempty"`
		Default   string
	}

	typ := reflect.TypeOf(sample{})

	sfCol, _ := typ.FieldByName("ColumnTag")
	require.Equal(t, "explicit", columnNameFromStructField(sfCol))

	sfJSON, _ := typ.FieldByName("JSONTag")
	require.Equal(t, "json_name", columnNameFromStructField(sfJSON))

	sfDefault, _ := typ.FieldByName("Default")
	require.Equal(t, "default", columnNameFromStructField(sfDefault))

	require.Equal(t, "explicit", parseColumnName("name:explicit"))
	require.Equal(t, "explicit", parseColumnName(" flag ; name:explicit ;"))
	require.Equal(t, "", parseColumnName(""))

	require.Equal(t, "name", tagPrimary("name,omitempty"))
	require.Equal(t, "", tagPrimary(""))
	require.Equal(t, "", tagPrimary("-"))
}

func TestQiEdgeCases(t *testing.T) {
	require.Equal(t, "\"\"", qi(""))
	require.Equal(t, "\"public\".\"users\"", qi("public.users"))
	require.Equal(t, "\"na\"\"me\"", qi("na\"me"))
	require.Equal(t, "*", qi("*"))
}

func TestSplitFunctionArgsAndParseFunctionCall(t *testing.T) {
	parts := splitFunctionArgs("func_call(1, 2), 'a''b', nested(3,4)")
	require.Equal(t, []string{"func_call(1, 2)", "'a''b'", "nested(3,4)"}, parts)

	name, args := parseFunctionCall("current_setting('some.key', text)")
	require.Equal(t, "current_setting", name)
	require.Len(t, args, 2)
	require.Equal(t, argString, args[0].kind)
	require.Equal(t, "some.key", args[0].value)
	require.Equal(t, argIdentifier, args[1].kind)
	require.Equal(t, "text", args[1].value)

	name, args = parseFunctionCall("json_build_object('a', func_call(1,2), 'c', 'd''e')")
	require.Equal(t, "json_build_object", name)
	require.Len(t, args, 4)
	require.Equal(t, argString, args[0].kind)
	require.Equal(t, "a", args[0].value)
	require.Equal(t, argUnknown, args[1].kind)
	require.Equal(t, "func_call(1,2)", args[1].value)
	require.Equal(t, "d'e", args[3].value)
}

func TestOperandToBuilderExpVariants(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		allowID  bool
		expected Exp
	}{
		{"string", "'hello'", true, String("hello")},
		{"true", "TRUE", true, Bool(true)},
		{"false", "FALSE", true, Bool(false)},
		{"int", "42", true, Int64(42)},
		{"identifier", "schema.table", true, Ident("schema.table")},
		{"raw", "jsonb_build_object('a',1)", false, Raw("jsonb_build_object('a',1)")},
		{"auth", "auth.uid()", true, AuthUID()},
		{"now", "now()", true, Now()},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			exp, _, ok := operandToBuilderExp(tc.input, tc.allowID)
			require.True(t, ok)
			require.Equal(t, tc.expected, exp)
		})
	}

	exp, code, ok := operandToBuilderExp("current_setting('request.jwt.claims', text)", true)
	require.True(t, ok)
	require.Equal(t, CurrentSettingAs("request.jwt.claims", "text"), exp)
	require.Equal(t, "st.CurrentSettingAs(\"request.jwt.claims\", \"text\")", code)
}

func TestParseClauseVariants(t *testing.T) {
	clause, code, ok := parseClause("users.id = 42")
	require.True(t, ok)
	require.Equal(t, Eq("users.id", Int64(42)), clause)
	require.Equal(t, "st.Eq(\"users.id\", st.Int64(42))", code)

	clause, code, ok = parseClause("role = 'admin' OR NOT (active = TRUE)")
	require.True(t, ok)
	expected := Or(Eq("role", String("admin")), Not(Eq("active", Bool(true))))
	require.Equal(t, expected.String(), clause.String())
	require.Contains(t, code, "st.Or")

	clause, code, ok = parseClause("JSONB_EXISTS(data, 'flag')")
	require.False(t, ok)
	require.Equal(t, Clause("JSONB_EXISTS(data, 'flag')"), clause)
	require.Equal(t, "st.Clause(\"JSONB_EXISTS(data, 'flag')\")", code)
}

func TestStripStorageBucketFilterVariants(t *testing.T) {
	bucket := "avatars"
	bucketClause := StorageBucketClause(bucket)
	require.Equal(t, "", StripStorageBucketFilter(bucketClause.String(), bucket))

	clause := bucketClause.String() + " AND owner = auth.uid()"
	require.Equal(t, "owner = auth.uid()", StripStorageBucketFilter(clause, bucket))

	orClause := bucketClause.String() + " OR role = 'admin'"
	stripped := StripStorageBucketFilter(orClause, bucket)
	require.Contains(t, stripped, "role")
	require.Contains(t, stripped, "'admin'")

	unrelated := "role = 'admin'"
	unrelatedResult := StripStorageBucketFilter(unrelated, bucket)
	require.Contains(t, unrelatedResult, "role")
	require.Contains(t, unrelatedResult, "'admin'")
}

func TestTrimOuterParens(t *testing.T) {
	require.Equal(t, "inner", trimOuterParens("(( inner ))"))
	require.Equal(t, "", trimOuterParens("()"))
}
