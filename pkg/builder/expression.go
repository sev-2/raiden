package builder

import (
	"strconv"
	"strings"
)

// ============================================================================
// Rule DSL — clear names, safe defaults, and practical helpers for Postgres RLS
// ============================================================================
//
// - Exp: RHS values or trusted function snippets.
// - Clause: boolean expressions for USING / WITH CHECK.
// - Identifiers are safely double-quoted (schema.table.column -> "schema"."table"."column").
// - String values are single-quoted with embedded quotes doubled.
// - Cast() only allows simple type identifiers (letters/digits/_).
// - And/Or/Not skip empty clauses and minimize redundant parentheses.
// - Includes extras: LIKE/ILIKE, regex, arrays, JSONB, FTS, EXISTS, ranges, dates, etc.
// - Backward-compatible shims provided at bottom (Val/S/I64/F64/...).
//
// Remove the Policy DDL generator per request.

// ----------------------------------------------------------------------------
// Types & Sentinels
// ----------------------------------------------------------------------------

type Exp string    // RHS literal or trusted function/expression
type Clause string // boolean expression fragment

func (e Exp) String() string    { return string(e) }
func (c Clause) String() string { return string(c) }
func (c Clause) IsEmpty() bool  { return strings.TrimSpace(string(c)) == "" }
func (c Clause) Paren() Clause {
	if c.IsEmpty() {
		return ""
	}
	return Clause("(" + c.String() + ")")
}

var (
	True  Clause = "TRUE"
	False Clause = "FALSE"
)

// ----------------------------------------------------------------------------
// Literals / Values
// ----------------------------------------------------------------------------

// Raw allows trusted fragments (avoid passing user input here).
func Raw(s string) Exp { return Exp(s) }

// String quotes a Go string as a SQL literal ('...'), escaping internal quotes.
func String(s string) Exp { return Exp("'" + strings.ReplaceAll(s, `'`, `''`) + "'") }

// Int64 formats a signed 64-bit integer literal.
func Int64(n int64) Exp { return Exp(strconv.FormatInt(n, 10)) }

// Float64 formats a float using compact, lossless representation.
func Float64(f float64) Exp { return Exp(strconv.FormatFloat(f, 'g', -1, 64)) }

// Cast applies ::type if type is a simple identifier (letters/digits/_).
func Cast(rhs Exp, typ string) Exp {
	typ = strings.TrimSpace(typ)
	if !isSimpleIdent(typ) {
		return rhs
	}
	return Exp(rhs.String() + "::" + typ)
}

// UUID casts a string literal to uuid safely.
func UUID(s string) Exp { return Cast(String(s), "uuid") }

// ----------------------------------------------------------------------------
// Session / Common Fragments
// ----------------------------------------------------------------------------

func AuthUID() Exp     { return Raw("auth.uid()") }
func CurrentUser() Exp { return Raw("current_user") }
func SessionUser() Exp { return Raw("session_user") }

func CurrentSetting(key string) Exp {
	return Raw("current_setting(" + String(key).String() + ")")
}

// CurrentSettingAs casts current_setting(key) to a given type (if safe).
func CurrentSettingAs(key, typ string) Exp {
	typ = strings.TrimSpace(typ)
	if typ == "" || !isSimpleIdent(typ) {
		return CurrentSetting(key)
	}
	return Exp(CurrentSetting(key).String() + "::" + typ)
}

// ----------------------------------------------------------------------------
// Identifier Helpers
// ----------------------------------------------------------------------------

// qi safely quotes dotted identifiers: schema.table or column -> "schema"."table" / "column".
func qi(ident string) string {
	ident = strings.TrimSpace(ident)
	if ident == "" {
		return `""`
	}
	parts := strings.Split(ident, ".")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.ReplaceAll(p, `"`, `""`)
		if p == "*" {
			out = append(out, "*")
			continue
		}
		out = append(out, `"`+p+`"`)
	}
	return strings.Join(out, ".")
}

// Keep your pointer-based resolver if you have it.
func ColOf(model any, fieldPtr any) string { return colNameFromPtr(model, fieldPtr) }

// ----------------------------------------------------------------------------
// Predicates (core)
// ----------------------------------------------------------------------------

func Eq(col string, rhs Exp) Clause { return Clause(qi(col) + " = " + rhs.String()) }
func Ne(col string, rhs Exp) Clause { return Clause(qi(col) + " <> " + rhs.String()) }
func Gt(col string, rhs Exp) Clause { return Clause(qi(col) + " > " + rhs.String()) }
func Ge(col string, rhs Exp) Clause { return Clause(qi(col) + " >= " + rhs.String()) }
func Lt(col string, rhs Exp) Clause { return Clause(qi(col) + " < " + rhs.String()) }
func Le(col string, rhs Exp) Clause { return Clause(qi(col) + " <= " + rhs.String()) }

func IsNull(col string) Clause  { return Clause(qi(col) + " IS NULL") }
func NotNull(col string) Clause { return Clause(qi(col) + " IS NOT NULL") }
func IsTrue(col string) Clause  { return Clause(qi(col) + " IS TRUE") }
func IsFalse(col string) Clause { return Clause(qi(col) + " IS FALSE") }

// Readable sugar for Eq(col, String(s)).
func EqString(col, s string) Clause { return Eq(col, String(s)) }

// IN / NOT IN
func In(col string, items ...Exp) Clause {
	if len(items) == 0 {
		return False
	}
	ss := make([]string, len(items))
	for i, it := range items {
		ss[i] = it.String()
	}
	return Clause(qi(col) + " IN (" + strings.Join(ss, ", ") + ")")
}
func InStrings(col string, values ...string) Clause {
	items := make([]Exp, len(values))
	for i, v := range values {
		items[i] = String(v)
	}
	return In(col, items...)
}
func NotIn(col string, items ...Exp) Clause {
	if len(items) == 0 {
		return True // nothing excluded
	}
	ss := make([]string, len(items))
	for i, it := range items {
		ss[i] = it.String()
	}
	return Clause(qi(col) + " NOT IN (" + strings.Join(ss, ", ") + ")")
}

// Logical composition
func Not(c Clause) Clause {
	if c.IsEmpty() {
		return ""
	}
	return Clause("NOT " + c.Paren().String())
}
func And(cs ...Clause) Clause { return joinClauses("AND", cs...) }
func Or(cs ...Clause) Clause  { return joinClauses("OR", cs...) }

func joinClauses(op string, cs ...Clause) Clause {
	parts := make([]string, 0, len(cs))
	for _, c := range cs {
		s := strings.TrimSpace(c.String())
		if s == "" {
			continue
		}
		if strings.HasPrefix(s, "(") && strings.HasSuffix(s, ")") {
			parts = append(parts, s)
		} else {
			parts = append(parts, "("+s+")")
		}
	}
	if len(parts) == 0 {
		return ""
	}
	return Clause(strings.Join(parts, " "+op+" "))
}

// ----------------------------------------------------------------------------
// Convenience macros
// ----------------------------------------------------------------------------

func TenantMatch(col string, settingKey, cast string) Clause {
	return Eq(col, CurrentSettingAs(settingKey, cast))
}
func OwnerIsAuth(col string) Clause { return Eq(col, AuthUID()) }

// ----------------------------------------------------------------------------
// Extras: Literals & small utilities
// ----------------------------------------------------------------------------

func Bool(b bool) Exp {
	if b {
		return Exp("TRUE")
	}
	return Exp("FALSE")
}

func Date(s string) Exp        { return Cast(String(s), "date") }
func Timestamp(s string) Exp   { return Cast(String(s), "timestamp") }
func Timestamptz(s string) Exp { return Cast(String(s), "timestamptz") }
func Interval(s string) Exp    { return Cast(String(s), "interval") }
func Now() Exp                 { return Raw("now()") }

// Use an identifier as an expression (compare column-to-column).
func Ident(name string) Exp { return Exp(qi(name)) }

// Generic function call: Func("lower", Ident("name"))
func Func(name string, args ...Exp) Exp {
	if !isSimpleIdent(name) {
		return Raw(name) // advanced users: pass Raw(...) if needed
	}
	as := make([]string, len(args))
	for i, a := range args {
		as[i] = a.String()
	}
	return Exp(name + "(" + strings.Join(as, ", ") + ")")
}

// Optional clause builders (include only when condition is true / value not empty)
func If(cond bool, c Clause) Clause {
	if cond {
		return c
	}
	return ""
}
func IfString(notEmpty string, f func(string) Clause) Clause {
	if strings.TrimSpace(notEmpty) == "" {
		return ""
	}
	return f(notEmpty)
}

// ----------------------------------------------------------------------------
// LIKE / ILIKE / Regex
// ----------------------------------------------------------------------------

func escapeLike(s string, esc rune) string {
	escS := string(esc)
	s = strings.ReplaceAll(s, escS, escS+escS)
	s = strings.ReplaceAll(s, "%", escS+"%")
	s = strings.ReplaceAll(s, "_", escS+"_")
	return s
}

func Like(col, pattern string) Clause {
	p := escapeLike(pattern, '\\')
	return Clause(qi(col) + " LIKE " + String(p).String() + " ESCAPE '\\'")
}
func ILike(col, pattern string) Clause {
	p := escapeLike(pattern, '\\')
	return Clause(qi(col) + " ILIKE " + String(p).String() + " ESCAPE '\\'")
}
func NotLike(col, pattern string) Clause  { return Not(Like(col, pattern)) }
func NotILike(col, pattern string) Clause { return Not(ILike(col, pattern)) }

func StartsWith(col, prefix string) Clause   { return Like(col, prefix+"%") }
func EndsWith(col, suffix string) Clause     { return Like(col, "%"+suffix) }
func ContainsText(col, needle string) Clause { return Like(col, "%"+needle+"%") }

func Regex(col, re string) Clause  { return Clause(qi(col) + " ~ " + String(re).String()) }
func IRegex(col, re string) Clause { return Clause(qi(col) + " ~* " + String(re).String()) }

// ----------------------------------------------------------------------------
// BETWEEN / DISTINCT
// ----------------------------------------------------------------------------

func Between(col string, low, high Exp) Clause {
	return Clause(qi(col) + " BETWEEN " + low.String() + " AND " + high.String())
}
func NotBetween(col string, low, high Exp) Clause {
	return Clause(qi(col) + " NOT BETWEEN " + low.String() + " AND " + high.String())
}

func IsDistinctFrom(col string, rhs Exp) Clause {
	return Clause(qi(col) + " IS DISTINCT FROM " + rhs.String())
}
func IsNotDistinctFrom(col string, rhs Exp) Clause {
	return Clause(qi(col) + " IS NOT DISTINCT FROM " + rhs.String())
}

// ----------------------------------------------------------------------------
// Arrays
// ----------------------------------------------------------------------------

func Array(items ...Exp) Exp {
	ss := make([]string, len(items))
	for i, it := range items {
		ss[i] = it.String()
	}
	return Exp("ARRAY[" + strings.Join(ss, ", ") + "]")
}
func ArrayStrings(values ...string) Exp {
	items := make([]Exp, len(values))
	for i, v := range values {
		items[i] = String(v)
	}
	return Array(items...)
}

func EqAny(col string, arr Exp) Clause { return Clause(qi(col) + " = ANY(" + arr.String() + ")") }
func EqAll(col string, arr Exp) Clause { return Clause(qi(col) + " = ALL(" + arr.String() + ")") }

func ArrayContains(col string, arr Exp) Clause    { return Clause(qi(col) + " @> " + arr.String()) }
func ArrayContainedBy(col string, arr Exp) Clause { return Clause(qi(col) + " <@ " + arr.String()) }
func ArrayOverlaps(col string, arr Exp) Clause    { return Clause(qi(col) + " && " + arr.String()) }

// ----------------------------------------------------------------------------
// JSON / JSONB
// ----------------------------------------------------------------------------

func JSONB(s string) Exp { return Exp(String(s).String() + "::jsonb") }

func JSONGet(col, key string) Exp     { return Exp(qi(col) + "->" + String(key).String()) }
func JSONGetText(col, key string) Exp { return Exp(qi(col) + "->>" + String(key).String()) }

func JSONPath(col string, path ...string) Exp {
	parts := make([]string, len(path))
	for i, p := range path {
		parts[i] = String(p).String()
	}
	return Exp(qi(col) + " #> ARRAY[" + strings.Join(parts, ", ") + "]")
}
func JSONPathText(col string, path ...string) Exp {
	parts := make([]string, len(path))
	for i, p := range path {
		parts[i] = String(p).String()
	}
	return Exp(qi(col) + " #>> ARRAY[" + strings.Join(parts, ", ") + "]")
}

func JSONContains(col string, json Exp) Clause { return Clause(qi(col) + " @> " + json.String()) }
func JSONHasKey(col, key string) Clause        { return Clause(qi(col) + " ? " + String(key).String()) }

func JSONHasAnyKeys(col string, keys ...string) Clause {
	if len(keys) == 0 {
		return False
	}
	ks := make([]string, len(keys))
	for i, k := range keys {
		ks[i] = String(k).String()
	}
	return Clause(qi(col) + " ?| ARRAY[" + strings.Join(ks, ", ") + "]")
}
func JSONHasAllKeys(col string, keys ...string) Clause {
	if len(keys) == 0 {
		return True
	}
	ks := make([]string, len(keys))
	for i, k := range keys {
		ks[i] = String(k).String()
	}
	return Clause(qi(col) + " ?& ARRAY[" + strings.Join(ks, ", ") + "]")
}

// ----------------------------------------------------------------------------
// Full-text search (FTS)
// ----------------------------------------------------------------------------

func TsVector(doc Exp, config ...string) Exp {
	if len(config) > 0 && isSimpleIdent(config[0]) {
		return Func("to_tsvector", Exp(config[0]), doc)
	}
	return Func("to_tsvector", doc)
}
func WebSearchQuery(q string, config ...string) Exp {
	if len(config) > 0 && isSimpleIdent(config[0]) {
		return Exp("websearch_to_tsquery(" + Exp(config[0]).String() + ", " + String(q).String() + ")")
	}
	return Exp("websearch_to_tsquery(" + String(q).String() + ")")
}
func TsMatches(vector Exp, query Exp) Clause {
	return Clause(vector.String() + " @@ " + query.String())
}

func Search(col, q, config string) Clause {
	return TsMatches(TsVector(Ident(col), config), WebSearchQuery(q, config))
}

// ----------------------------------------------------------------------------
// EXISTS helpers
// ----------------------------------------------------------------------------

func Exists(rawSelectSQL string) Clause {
	sql := strings.TrimSpace(rawSelectSQL)
	if !strings.HasPrefix(strings.ToUpper(sql), "SELECT ") {
		sql = "SELECT " + sql // allow "1 FROM ..." shorthand
	}
	return Clause("EXISTS (" + sql + ")")
}
func NotExists(rawSelectSQL string) Clause { return Not(Exists(rawSelectSQL)) }

// Common pattern: EXISTS (SELECT 1 FROM schema.table t WHERE joinCond AND extra...)
func ExistsFrom(table string, joinCond Clause, extra ...Clause) Clause {
	where := And(joinCond, And(extra...))
	return Clause("EXISTS (SELECT 1 FROM " + qi(table) + " WHERE " + where.String() + ")")
}

// ----------------------------------------------------------------------------
// Ranges (tsrange/int4range/etc.)
// ----------------------------------------------------------------------------

func RangeContains(col string, rhs Exp) Clause    { return Clause(qi(col) + " @> " + rhs.String()) }
func RangeContainedBy(col string, rhs Exp) Clause { return Clause(qi(col) + " <@ " + rhs.String()) }
func RangeOverlaps(col string, rhs Exp) Clause    { return Clause(qi(col) + " && " + rhs.String()) }

// ----------------------------------------------------------------------------
// Null handling
// ----------------------------------------------------------------------------

func Coalesce(args ...Exp) Exp {
	if len(args) == 0 {
		return Raw("NULL")
	}
	parts := make([]string, len(args))
	for i, a := range args {
		parts[i] = a.String()
	}
	return Exp("COALESCE(" + strings.Join(parts, ", ") + ")")
}
func NullIf(a, b Exp) Exp { return Exp("NULLIF(" + a.String() + ", " + b.String() + ")") }

// ----------------------------------------------------------------------------
// Internals
// ----------------------------------------------------------------------------

func isSimpleIdent(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !(r == '_' || r >= '0' && r <= '9' || r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z') {
			return false
		}
	}
	return true
}

// ----------------------------------------------------------------------------
// Backward-compatible shims (keep for now; remove later)
// ----------------------------------------------------------------------------

// Deprecated: use String.
func Val(s string) Exp { return String(s) }

// Deprecated: use String.
func S(s string) Exp { return String(s) }

// Deprecated: use Int64.
func I64(n int64) Exp { return Int64(n) }

// Deprecated: use Float64.
func F64(f float64) Exp { return Float64(f) }

// Deprecated: use UUID.
func UUIDLit(s string) Exp { return UUID(s) }

// Deprecated: use EqString.
func EqS(col, s string) Clause { return EqString(col, s) }

// Deprecated: use InStrings.
func InS(col string, values ...string) Clause { return InStrings(col, values...) }

// Deprecated: use CurrentSettingAs.
func CurrentSettingCast(key, cast string) Exp { return CurrentSettingAs(key, cast) }
