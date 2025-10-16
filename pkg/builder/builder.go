package builder

import (
	"reflect"
	"strings"

	"github.com/jinzhu/inflection"
	"github.com/sev-2/raiden/pkg/utils"
)

// ============================================================================
// Tag readers (schema/table; &field -> column)
// ============================================================================

func TableFromModel(model any) (schema, table string) {
	schema = "public"
	table = ""

	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() == reflect.Struct {
		for i := 0; i < t.NumField(); i++ {
			sf := t.Field(i)
			if v := sf.Tag.Get("schema"); v != "" {
				schema = v
			}
			if v := sf.Tag.Get("tableName"); v != "" {
				table = v
			}
		}
		if table == "" {
			table = inflection.Plural(utils.ToSnakeCase(t.Name()))
		}
	} else {
		table = "unknown"
	}
	return schema, table
}

func colNameFromPtr(model any, fieldPtr any) string {
	if model == nil || fieldPtr == nil {
		panic("builder: model and fieldPtr are required")
	}

	mv := reflect.ValueOf(model)
	if mv.Kind() != reflect.Ptr || mv.IsNil() {
		panic("builder: model must be a non-nil pointer to struct")
	}
	mv = mv.Elem()
	if mv.Kind() != reflect.Struct {
		panic("builder: model must point to a struct")
	}

	fp := reflect.ValueOf(fieldPtr)
	if fp.Kind() != reflect.Ptr || fp.IsNil() {
		panic("builder: fieldPtr must be a pointer to a struct field (e.g., &m.Title)")
	}
	target := fp.Pointer()

	mt := mv.Type()
	for i := 0; i < mv.NumField(); i++ {
		fv := mv.Field(i)
		if !fv.CanAddr() {
			continue
		}
		if fv.Addr().Pointer() == target {
			sf := mt.Field(i)

			// 1) column tag: name:...
			if col := parseColumnName(sf.Tag.Get("column")); col != "" {
				return col
			}
			// 2) json tag
			if col := tagPrimary(sf.Tag.Get("json")); col != "" {
				return col
			}
			// 3) fallback: snake_case(FieldName)
			return utils.ToSnakeCase(sf.Name)
		}
	}

	panic("builder: could not match fieldPtr to any addressable struct field")
}

func parseColumnName(colTag string) string {
	if colTag == "" {
		return ""
	}
	for _, p := range strings.Split(colTag, ";") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if strings.HasPrefix(p, "name:") {
			return strings.TrimSpace(strings.TrimPrefix(p, "name:"))
		}
	}
	return ""
}

func tagPrimary(tag string) string {
	if tag == "" || tag == "-" {
		return ""
	}
	if i := strings.Index(tag, ","); i >= 0 {
		return tag[:i]
	}
	return tag
}
