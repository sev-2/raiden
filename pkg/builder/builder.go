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
	for mv.Kind() == reflect.Pointer {
		if mv.IsNil() {
			mv = reflect.New(mv.Type().Elem())
		}
		mv = mv.Elem()
	}

	if mv.Kind() != reflect.Struct {
		panic(fmt.Sprintf("builder: model must resolve to struct (got %T)", model))
	}

	if !mv.CanAddr() {
		addr := reflect.New(mv.Type())
		addr.Elem().Set(mv)
		mv = addr.Elem()
	}

	fp := reflect.ValueOf(fieldPtr)
	if !fp.IsValid() {
		panic("builder: fieldPtr must not be nil")
	}

	mt := mv.Type()

	if fp.Kind() == reflect.Pointer && !fp.IsNil() {
		target := fp.Pointer()
		bestMatch, fallbackMatch := resolveColumnFromStruct(mv, mt, target)
		if bestMatch != nil && bestMatch.score > 0 {
			return bestMatch.column
		}

		// Attempt to resolve by matching the field type when address comparison fails
		fieldType := fp.Type().Elem()
		if column, ok := lookupColumnByType(mt, fieldType); ok {
			return column
		}

		if fallbackMatch != nil && fallbackMatch.column != "" {
			return fallbackMatch.column
		}

		panic("builder: could not match fieldPtr to any addressable struct field")
	}

	// Allow resolving by value or nil pointer (e.g., model.Owner where Owner is *uuid.UUID and nil)
	fieldType := fp.Type()

	if fp.Kind() == reflect.Pointer && fp.IsNil() {
		// keep fieldType as the pointer type, matching the struct field declaration
		if column, ok := lookupColumnByType(mt, fieldType); ok {
			return column
		}
		panic("builder: fieldPtr must be a pointer to a struct field (e.g., &m.Title)")
	}

	if column, ok := lookupColumnByType(mt, fieldType); ok {
		return column
	}

	panic("builder: fieldPtr must be a pointer to a struct field (e.g., &m.Title)")
}

func lookupColumnByType(t reflect.Type, fieldType reflect.Type) (string, bool) {
	var best *fieldMatch
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if sf.Type != fieldType {
			continue
		}

		score := fieldMatchScore(sf)
		candidate := &fieldMatch{column: columnNameFromStructField(sf), score: score}
		if best == nil || candidate.score > best.score {
			best = candidate
		}
	}

	if best != nil {
		return best.column, true
	}

	return "", false
}

type fieldMatch struct {
	column string
	score  int
}

func resolveColumnFromStruct(v reflect.Value, t reflect.Type, target uintptr) (best *fieldMatch, fallback *fieldMatch) {
	updateBest := func(candidate *fieldMatch) {
		if candidate == nil {
			return
		}
		if best == nil || candidate.score > best.score {
			best = candidate
		}
	}

	updateFallback := func(candidate *fieldMatch) {
		if candidate == nil {
			return
		}
		if fallback == nil {
			fallback = candidate
		}
	}

	for i := 0; i < v.NumField(); i++ {
		fv := v.Field(i)
		sf := t.Field(i)

		var candidate *fieldMatch
		if fv.CanAddr() && fv.Addr().Pointer() == target {
			score := fieldMatchScore(sf)
			candidate = &fieldMatch{column: columnNameFromStructField(sf), score: score}
			if score <= 0 {
				updateFallback(candidate)
			}
		}

		switch fv.Kind() {
		case reflect.Struct:
			if nestedBest, nestedFallback := resolveColumnFromStruct(fv, sf.Type, target); nestedBest != nil || nestedFallback != nil {
				updateBest(nestedBest)
				if nestedBest == nil {
					updateFallback(nestedFallback)
				}
			}
		case reflect.Ptr:
			if !fv.IsNil() && fv.Elem().Kind() == reflect.Struct {
				if nestedBest, nestedFallback := resolveColumnFromStruct(fv.Elem(), fv.Elem().Type(), target); nestedBest != nil || nestedFallback != nil {
					updateBest(nestedBest)
					if nestedBest == nil {
						updateFallback(nestedFallback)
					}
				}
			}
		}

		if candidate != nil {
			if candidate.score > 0 {
				updateBest(candidate)
			} else {
				updateFallback(candidate)
			}
		}
	}

	return best, fallback
}

func fieldMatchScore(sf reflect.StructField) int {
	score := 0
	columnTag := strings.TrimSpace(sf.Tag.Get("column"))
	jsonTag := strings.TrimSpace(sf.Tag.Get("json"))

	if columnTag != "" {
		score += 200
	}
	if jsonTag != "" && jsonTag != "-" {
		score += 50
	}
	if !sf.Anonymous {
		score += 10
	}
	if score == 0 && columnTag == "" && !sf.Anonymous {
		score = 1
	}
	return score
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
