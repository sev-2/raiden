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
	bestMatch, fallbackMatch := resolveColumnFromStruct(mv, mt, target)
	if bestMatch != nil && bestMatch.score > 0 {
		return bestMatch.column
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

	if fallbackMatch != nil && fallbackMatch.column != "" {
		return fallbackMatch.column
	}

	panic("builder: could not match fieldPtr to any addressable struct field")
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
