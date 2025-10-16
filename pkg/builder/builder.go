package builder

import (
	"fmt"
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
	switch mv.Kind() {
	case reflect.Ptr:
		if mv.IsNil() {
			// create zero value for metadata purposes
			mv = reflect.New(mv.Type().Elem())
		}
		mv = mv.Elem()
	case reflect.Struct:
		// make addressable copy so we can inspect metadata
		addr := reflect.New(mv.Type())
		addr.Elem().Set(mv)
		mv = addr.Elem()
	default:
		panic(fmt.Sprintf("builder: model must be struct or pointer to struct (got %T)", model))
	}

	if mv.Kind() != reflect.Struct {
		panic(fmt.Sprintf("builder: model must resolve to struct (got %T)", model))
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
			return columnNameFromStructField(mt.Field(i))
		}
	}

	// Fallback: attempt to resolve by matching field type
	fieldType := fp.Type().Elem()
	var candidates []reflect.StructField
	for i := 0; i < mt.NumField(); i++ {
		sf := mt.Field(i)
		if sf.Type == fieldType {
			candidates = append(candidates, sf)
		}
	}

	if len(candidates) == 1 {
		return columnNameFromStructField(candidates[0])
	}

	panic("builder: could not match fieldPtr to any addressable struct field")
}

func columnNameFromStructField(sf reflect.StructField) string {
	if col := parseColumnName(sf.Tag.Get("column")); col != "" {
		return col
	}
	if col := tagPrimary(sf.Tag.Get("json")); col != "" {
		return col
	}
	return utils.ToSnakeCase(sf.Name)
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
		if after, ok := strings.CutPrefix(p, "name:"); ok {
			return strings.TrimSpace(after)
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
