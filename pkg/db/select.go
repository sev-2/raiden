package db

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/sev-2/raiden"
)

func (q Query) Select(columns []string, aliases map[string]string) (model *Query) {

	for _, column := range columns {
		if !isValidColumn(q.model, column) {
			errorMessage := fmt.Sprintf(
				"invalid column: \"%s\" is not available on \"%s\" table.",
				column,
				GetTable(q.model),
			)

			raiden.Fatal(errorMessage)
		}
	}

	for column, _ := range aliases {
		if !isValidColumn(q.model, column) {
			errorMessage := fmt.Sprintf(
				"invalid alias column: \"%s\" is not available on \"%s\" table.",
				column,
				GetTable(q.model),
			)

			raiden.Fatal(errorMessage)
		}
	}

	for i, column := range columns {
		if aliases[column] != "" {
			columns[i] = aliases[column] + ":" + column
		}
	}

	q.Columns = columns
	return &q
}

func GetColumnList(m interface{}) []string {
	r := reflect.TypeOf(m)

	if r.Kind() == reflect.Ptr {
		r = r.Elem()
	}

	var columns []string

	for i := 0; i < r.NumField(); i++ {
		field := r.Field(i)
		tag := field.Tag.Get("column")
		if tag != "" {
			for _, t := range strings.Split(tag, ";") {
				if strings.HasPrefix(t, "name:") {
					name := strings.TrimPrefix(t, "name:")
					columns = append(columns, name)
				}
			}
		}
	}

	return columns
}

func isValidColumn(model interface{}, column string) bool {
	if column == "*" {
		return true
	}

	validSet := make(map[string]bool)

	for _, v := range GetColumnList(model) {
		validSet[v] = true
	}

	if validSet[column] {
		return true
	}

	return false
}
