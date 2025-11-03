package builder_test

import (
	"testing"

	"github.com/sev-2/raiden/pkg/builder"
	"github.com/stretchr/testify/assert"
)

// Test Int64 function
func TestInt64(t *testing.T) {
	result := builder.Int64(123)
	assert.Equal(t, builder.Exp("123"), result)
}

// Test Float64 function
func TestFloat64(t *testing.T) {
	result := builder.Float64(123.45)
	assert.Equal(t, builder.Exp("123.45"), result)
}

// Test Cast function
func TestCast(t *testing.T) {
	t.Run("valid type", func(t *testing.T) {
		result := builder.Cast(builder.String("test"), "text")
		assert.Equal(t, builder.Exp("'test'::text"), result)
	})

	t.Run("invalid type", func(t *testing.T) {
		result := builder.Cast(builder.String("test"), "text;DROP TABLE users")
		assert.Equal(t, builder.Exp("'test'"), result) // Should not cast due to invalid type
	})

	t.Run("empty type", func(t *testing.T) {
		result := builder.Cast(builder.String("test"), "")
		assert.Equal(t, builder.Exp("'test'"), result)
	})
}

// Test UUID function
func TestUUID(t *testing.T) {
	result := builder.UUID("123e4567-e89b-12d3-a456-426614174000")
	assert.Equal(t, builder.Exp("'123e4567-e89b-12d3-a456-426614174000'::uuid"), result)
}

// Test CurrentUser function
func TestCurrentUser(t *testing.T) {
	result := builder.CurrentUser()
	assert.Equal(t, builder.Exp("current_user"), result)
}

// Test SessionUser function
func TestSessionUser(t *testing.T) {
	result := builder.SessionUser()
	assert.Equal(t, builder.Exp("session_user"), result)
}

// Test CurrentSetting function
func TestCurrentSetting(t *testing.T) {
	result := builder.CurrentSetting("some.setting")
	assert.Equal(t, builder.Exp("current_setting('some.setting')"), result)
}

// Test CurrentSettingAs function
func TestCurrentSettingAs(t *testing.T) {
	t.Run("valid type", func(t *testing.T) {
		result := builder.CurrentSettingAs("some.setting", "text")
		assert.Equal(t, builder.Exp("current_setting('some.setting')::text"), result)
	})

	t.Run("invalid type", func(t *testing.T) {
		result := builder.CurrentSettingAs("some.setting", "text;DROP TABLE users")
		assert.Equal(t, builder.Exp("current_setting('some.setting')"), result)
	})
}

// Test Ne function
func TestNe(t *testing.T) {
	result := builder.Ne("col", builder.String("value"))
	assert.Equal(t, builder.Clause(`"col" <> 'value'`), result)
}

// Test Gt function
func TestGt(t *testing.T) {
	result := builder.Gt("col", builder.Int64(5))
	assert.Equal(t, builder.Clause(`"col" > 5`), result)
}

// Test Ge function
func TestGe(t *testing.T) {
	result := builder.Ge("col", builder.Int64(5))
	assert.Equal(t, builder.Clause(`"col" >= 5`), result)
}

// Test Lt function
func TestLt(t *testing.T) {
	result := builder.Lt("col", builder.Int64(5))
	assert.Equal(t, builder.Clause(`"col" < 5`), result)
}

// Test Le function
func TestLe(t *testing.T) {
	result := builder.Le("col", builder.Int64(5))
	assert.Equal(t, builder.Clause(`"col" <= 5`), result)
}

// Test IsNull function
func TestIsNull(t *testing.T) {
	result := builder.IsNull("col")
	assert.Equal(t, builder.Clause(`"col" IS NULL`), result)
}

// Test NotNull function
func TestNotNull(t *testing.T) {
	result := builder.NotNull("col")
	assert.Equal(t, builder.Clause(`"col" IS NOT NULL`), result)
}

// Test IsTrue function
func TestIsTrue(t *testing.T) {
	result := builder.IsTrue("col")
	assert.Equal(t, builder.Clause(`"col" IS TRUE`), result)
}

// Test IsFalse function
func TestIsFalse(t *testing.T) {
	result := builder.IsFalse("col")
	assert.Equal(t, builder.Clause(`"col" IS FALSE`), result)
}

// Test In function
func TestIn(t *testing.T) {
	t.Run("with items", func(t *testing.T) {
		result := builder.In("col", builder.String("a"), builder.String("b"))
		assert.Equal(t, builder.Clause(`"col" IN ('a', 'b')`), result)
	})

	t.Run("with no items", func(t *testing.T) {
		result := builder.In("col")
		assert.Equal(t, builder.False, result)
	})
}

// Test InStrings function
func TestInStrings(t *testing.T) {
	result := builder.InStrings("col", "a", "b")
	assert.Equal(t, builder.Clause(`"col" IN ('a', 'b')`), result)
}

// Test NotIn function
func TestNotIn(t *testing.T) {
	t.Run("with items", func(t *testing.T) {
		result := builder.NotIn("col", builder.String("a"), builder.String("b"))
		assert.Equal(t, builder.Clause(`"col" NOT IN ('a', 'b')`), result)
	})

	t.Run("with no items", func(t *testing.T) {
		result := builder.NotIn("col")
		assert.Equal(t, builder.True, result) // nothing excluded
	})
}

// Test Not function
func TestNot(t *testing.T) {
	t.Run("with clause", func(t *testing.T) {
		result := builder.Not(builder.Eq("col", builder.String("value")))
		assert.Equal(t, builder.Clause("NOT (\"col\" = 'value')"), result)
	})

	t.Run("with empty clause", func(t *testing.T) {
		result := builder.Not("")
		assert.Equal(t, builder.Clause(""), result)
	})
}

// Test Or function
func TestOr(t *testing.T) {
	result := builder.Or(builder.Eq("col1", builder.String("value1")), builder.Eq("col2", builder.String("value2")))
	assert.Contains(t, result.String(), "(")
	assert.Contains(t, result.String(), "OR")
}

// Test TenantMatch function
func TestTenantMatch(t *testing.T) {
	result := builder.TenantMatch("col", "tenant_id", "uuid")
	assert.Contains(t, result.String(), "col")
	assert.Contains(t, result.String(), "tenant_id")
}

func TestTsVectorVariants(t *testing.T) {
	base := builder.TsVector(builder.Ident("content"))
	assert.Equal(t, "to_tsvector(\"content\")", base.String())

	withConfig := builder.TsVector(builder.Ident("content"), "english")
	assert.Equal(t, "to_tsvector(english, \"content\")", withConfig.String())

	invalidConfig := builder.TsVector(builder.Ident("content"), "pg.simple")
	assert.Equal(t, base.String(), invalidConfig.String())
}

func TestWebSearchQueryVariants(t *testing.T) {
	defaultQuery := builder.WebSearchQuery("hello world")
	assert.Equal(t, "websearch_to_tsquery('hello world')", defaultQuery.String())

	withConfig := builder.WebSearchQuery("hello", "english")
	assert.Equal(t, "websearch_to_tsquery(english, 'hello')", withConfig.String())

	invalidConfig := builder.WebSearchQuery("hello world", "pg.simple")
	assert.Equal(t, defaultQuery.String(), invalidConfig.String())
}

// Test OwnerIsAuth function
func TestOwnerIsAuth(t *testing.T) {
	result := builder.OwnerIsAuth("col")
	assert.Contains(t, result.String(), "col")
	assert.Contains(t, result.String(), "auth.uid()")
}

func TestExistsHelpers(t *testing.T) {
	plain := builder.Exists("1 FROM data")
	assert.Equal(t, "EXISTS (SELECT 1 FROM data)", plain.String())

	notExists := builder.NotExists("SELECT 1 FROM data")
	assert.Equal(t, "NOT (EXISTS (SELECT 1 FROM data))", notExists.String())

	from := builder.ExistsFrom("public.files", builder.Eq("files.bucket_id", builder.String("avatars")), builder.Eq("files.owner", builder.AuthUID()))
	assert.Contains(t, from.String(), "EXISTS (SELECT 1 FROM \"public\".\"files\" WHERE")
	assert.Contains(t, from.String(), "auth.uid()")
}

func TestBackwardCompatibilityHelpers(t *testing.T) {
	u := builder.UUIDLit("123e4567-e89b-12d3-a456-426614174000")
	assert.Equal(t, builder.UUID("123e4567-e89b-12d3-a456-426614174000"), u)

	setting := builder.CurrentSettingCast("my.key", "text")
	assert.Equal(t, builder.CurrentSettingAs("my.key", "text"), setting)
}

func TestJwtRoleHelpers(t *testing.T) {
	roleClause := builder.RoleIs("editor")
	assert.Contains(t, roleClause.String(), "'role'")
	assert.Contains(t, roleClause.String(), "'editor'")

	rolesClause := builder.RolesAny("admin", "editor")
	assert.Contains(t, rolesClause.String(), "ANY(ARRAY['admin', 'editor'])")

	inOrg := builder.InOrg("files.org_id")
	assert.Equal(t, builder.Eq("files.org_id", builder.Claim("org_id")), inOrg)

	assert.Equal(t, builder.CurrentSetting("claim"), builder.Claim("claim"))
}

// Test Bool function
func TestBool(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		result := builder.Bool(true)
		assert.Equal(t, builder.Exp("TRUE"), result)
	})

	t.Run("false", func(t *testing.T) {
		result := builder.Bool(false)
		assert.Equal(t, builder.Exp("FALSE"), result)
	})
}

// Test Date function
func TestDate(t *testing.T) {
	result := builder.Date("2023-01-01")
	assert.Equal(t, builder.Exp("'2023-01-01'::date"), result)
}

// Test Timestamp function
func TestTimestamp(t *testing.T) {
	result := builder.Timestamp("2023-01-01 12:00:00")
	assert.Equal(t, builder.Exp("'2023-01-01 12:00:00'::timestamp"), result)
}

// Test Timestamptz function
func TestTimestamptz(t *testing.T) {
	result := builder.Timestamptz("2023-01-01 12:00:00")
	assert.Equal(t, builder.Exp("'2023-01-01 12:00:00'::timestamptz"), result)
}

// Test Interval function
func TestInterval(t *testing.T) {
	result := builder.Interval("1 day")
	assert.Equal(t, builder.Exp("'1 day'::interval"), result)
}

// Test Now function
func TestNow(t *testing.T) {
	result := builder.Now()
	assert.Equal(t, builder.Exp("now()"), result)
}

// Test Ident function
func TestIdent(t *testing.T) {
	result := builder.Ident("col.name")
	assert.Equal(t, builder.Exp(`"col"."name"`), result)
}

// Test Func function
func TestFunc(t *testing.T) {
	t.Run("valid function name", func(t *testing.T) {
		result := builder.Func("lower", builder.Ident("name"))
		assert.Equal(t, builder.Exp("lower(\"name\")"), result)
	})

	t.Run("invalid function name", func(t *testing.T) {
		result := builder.Func("lower;DROP TABLE users", builder.Ident("name"))
		assert.Equal(t, builder.Exp("lower;DROP TABLE users"), result)
	})
}

// Test If function
func TestIf(t *testing.T) {
	t.Run("condition true", func(t *testing.T) {
		clause := builder.Eq("col", builder.String("value"))
		result := builder.If(true, clause)
		assert.Equal(t, clause, result)
	})

	t.Run("condition false", func(t *testing.T) {
		clause := builder.Eq("col", builder.String("value"))
		result := builder.If(false, clause)
		assert.Equal(t, builder.Clause(""), result)
	})
}

// Test IfString function
func TestIfString(t *testing.T) {
	t.Run("non-empty string", func(t *testing.T) {
		result := builder.IfString("test", func(s string) builder.Clause {
			return builder.Eq("col", builder.String(s))
		})
		assert.Equal(t, builder.Clause(`"col" = 'test'`), result)
	})

	t.Run("empty string", func(t *testing.T) {
		result := builder.IfString("", func(s string) builder.Clause {
			return builder.Eq("col", builder.String(s))
		})
		assert.Equal(t, builder.Clause(""), result)
	})
}

// Test Like functions
func TestLike(t *testing.T) {
	result := builder.Like("col", "pattern%")
	assert.Contains(t, result.String(), "LIKE")
}

func TestILike(t *testing.T) {
	result := builder.ILike("col", "pattern%")
	assert.Contains(t, result.String(), "ILIKE")
}

func TestNotLike(t *testing.T) {
	result := builder.NotLike("col", "pattern%")
	assert.Contains(t, result.String(), "NOT")
	assert.Contains(t, result.String(), "LIKE")
}

func TestNotILike(t *testing.T) {
	result := builder.NotILike("col", "pattern%")
	assert.Contains(t, result.String(), "NOT")
	assert.Contains(t, result.String(), "ILIKE")
}

func TestStartsWith(t *testing.T) {
	result := builder.StartsWith("col", "prefix")
	assert.Contains(t, result.String(), "LIKE")
	assert.Contains(t, result.String(), "prefix")
}

func TestEndsWith(t *testing.T) {
	result := builder.EndsWith("col", "suffix")
	assert.Contains(t, result.String(), "LIKE")
	assert.Contains(t, result.String(), "suffix")
}

func TestContainsText(t *testing.T) {
	result := builder.ContainsText("col", "text")
	assert.Contains(t, result.String(), "LIKE")
	assert.Contains(t, result.String(), "text")
}

// Test Regex functions
func TestRegex(t *testing.T) {
	result := builder.Regex("col", "[a-z]+")
	assert.Contains(t, result.String(), "~")
}

func TestIRegex(t *testing.T) {
	result := builder.IRegex("col", "[a-z]+")
	assert.Contains(t, result.String(), "~*")
}

// Test Between functions
func TestBetween(t *testing.T) {
	result := builder.Between("col", builder.Int64(1), builder.Int64(10))
	assert.Contains(t, result.String(), "BETWEEN")
}

func TestNotBetween(t *testing.T) {
	result := builder.NotBetween("col", builder.Int64(1), builder.Int64(10))
	assert.Contains(t, result.String(), "NOT BETWEEN")
}

// Test Distinct functions
func TestIsDistinctFrom(t *testing.T) {
	result := builder.IsDistinctFrom("col", builder.String("value"))
	assert.Contains(t, result.String(), "IS DISTINCT FROM")
}

func TestIsNotDistinctFrom(t *testing.T) {
	result := builder.IsNotDistinctFrom("col", builder.String("value"))
	assert.Contains(t, result.String(), "IS NOT DISTINCT FROM")
}

// Test Array functions
func TestArray(t *testing.T) {
	result := builder.Array(builder.String("a"), builder.String("b"))
	assert.Contains(t, result.String(), "ARRAY")
}

func TestArrayStrings(t *testing.T) {
	result := builder.ArrayStrings("a", "b")
	assert.Contains(t, result.String(), "ARRAY")
}

// Test Array operators
func TestEqAny(t *testing.T) {
	arr := builder.Array(builder.String("a"), builder.String("b"))
	result := builder.EqAny("col", arr)
	assert.Contains(t, result.String(), "= ANY")
}

func TestEqAll(t *testing.T) {
	arr := builder.Array(builder.String("a"), builder.String("b"))
	result := builder.EqAll("col", arr)
	assert.Contains(t, result.String(), "= ALL")
}

func TestArrayContains(t *testing.T) {
	arr := builder.Array(builder.String("a"), builder.String("b"))
	result := builder.ArrayContains("col", arr)
	assert.Contains(t, result.String(), "@>")
}

func TestArrayContainedBy(t *testing.T) {
	arr := builder.Array(builder.String("a"), builder.String("b"))
	result := builder.ArrayContainedBy("col", arr)
	assert.Contains(t, result.String(), "<@")
}

func TestArrayOverlaps(t *testing.T) {
	arr := builder.Array(builder.String("a"), builder.String("b"))
	result := builder.ArrayOverlaps("col", arr)
	assert.Contains(t, result.String(), "&&")
}

// Test JSON functions
func TestJSONB(t *testing.T) {
	result := builder.JSONB(`{"key": "value"}`)
	assert.Contains(t, result.String(), "::jsonb")
}

func TestJSONGet(t *testing.T) {
	result := builder.JSONGet("col", "key")
	assert.Contains(t, result.String(), "->")
}

func TestJSONGetText(t *testing.T) {
	result := builder.JSONGetText("col", "key")
	assert.Contains(t, result.String(), "->>")
}

func TestJSONPath(t *testing.T) {
	result := builder.JSONPath("col", "key", "subkey")
	assert.Contains(t, result.String(), "#>")
}

func TestJSONPathText(t *testing.T) {
	result := builder.JSONPathText("col", "key", "subkey")
	assert.Contains(t, result.String(), "#>>")
}

func TestJSONContains(t *testing.T) {
	result := builder.JSONContains("col", builder.JSONB(`{"key": "value"}`))
	assert.Contains(t, result.String(), "@>")
}

func TestJSONHasKey(t *testing.T) {
	result := builder.JSONHasKey("col", "key")
	assert.Contains(t, result.String(), "?")
}

func TestJSONHasAnyKeys(t *testing.T) {
	result := builder.JSONHasAnyKeys("col", "key1", "key2")
	assert.Contains(t, result.String(), "?|")
}

func TestJSONHasAllKeys(t *testing.T) {
	result := builder.JSONHasAllKeys("col", "key1", "key2")
	assert.Contains(t, result.String(), "?&")
}

// Test FTS functions
func TestTsVector(t *testing.T) {
	result := builder.TsVector(builder.String("document"))
	assert.Contains(t, result.String(), "to_tsvector")
}

func TestWebSearchQuery(t *testing.T) {
	result := builder.WebSearchQuery("search term")
	assert.Contains(t, result.String(), "websearch_to_tsquery")
}

func TestTsMatches(t *testing.T) {
	vector := builder.TsVector(builder.String("document"))
	query := builder.WebSearchQuery("term")
	result := builder.TsMatches(vector, query)
	assert.Contains(t, result.String(), "@@")
}

func TestSearch(t *testing.T) {
	result := builder.Search("col", "query", "")
	assert.Contains(t, result.String(), "@@")
}

// Test EXISTS functions
func TestExists(t *testing.T) {
	result := builder.Exists("SELECT 1 FROM table WHERE col = 'value'")
	assert.Contains(t, result.String(), "EXISTS")
}

func TestNotExists(t *testing.T) {
	result := builder.NotExists("SELECT 1 FROM table WHERE col = 'value'")
	assert.Contains(t, result.String(), "NOT")
	assert.Contains(t, result.String(), "EXISTS")
}

func TestExistsFrom(t *testing.T) {
	joinClause := builder.Eq("t.id", builder.Int64(1))
	result := builder.ExistsFrom("table", joinClause)
	assert.Contains(t, result.String(), "EXISTS")
	assert.Contains(t, result.String(), "SELECT 1 FROM")
}

// Test Range functions
func TestRangeContains(t *testing.T) {
	result := builder.RangeContains("col", builder.Raw("range_value"))
	assert.Contains(t, result.String(), "@>")
}

func TestRangeContainedBy(t *testing.T) {
	result := builder.RangeContainedBy("col", builder.Raw("range_value"))
	assert.Contains(t, result.String(), "<@")
}

func TestRangeOverlaps(t *testing.T) {
	result := builder.RangeOverlaps("col", builder.Raw("range_value"))
	assert.Contains(t, result.String(), "&&")
}

// Test Coalesce function
func TestCoalesce(t *testing.T) {
	t.Run("with args", func(t *testing.T) {
		result := builder.Coalesce(builder.String("a"), builder.String("b"))
		assert.Contains(t, result.String(), "COALESCE")
	})

	t.Run("no args", func(t *testing.T) {
		result := builder.Coalesce()
		assert.Equal(t, builder.Raw("NULL"), result)
	})
}

// Test NullIf function
func TestNullIf(t *testing.T) {
	result := builder.NullIf(builder.String("a"), builder.String("b"))
	assert.Contains(t, result.String(), "NULLIF")
}

// Test deprecated functions
func TestDeprecatedFunctions(t *testing.T) {
	assert.Equal(t, builder.String("test"), builder.Val("test"))
	assert.Equal(t, builder.String("test"), builder.S("test"))
	assert.Equal(t, builder.Int64(123), builder.I64(123))
	assert.Equal(t, builder.Float64(123.45), builder.F64(123.45))
	assert.Equal(t, builder.EqString("col", "value"), builder.EqS("col", "value"))
	assert.Equal(t, builder.InStrings("col", "a", "b"), builder.InS("col", "a", "b"))
}

// Test Clause methods
func TestClauseParen(t *testing.T) {
	t.Run("with clause", func(t *testing.T) {
		clause := builder.Eq("col", builder.String("value"))
		result := clause.Paren()
		assert.Equal(t, builder.Clause("(\"col\" = 'value')"), result)
	})

	t.Run("with empty clause", func(t *testing.T) {
		clause := builder.Clause("")
		result := clause.Paren()
		assert.Equal(t, builder.Clause(""), result)
	})
}
