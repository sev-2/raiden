package db

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/sev-2/raiden"
)

func (q Query) Select(columns []string) (model *Query) {

	for _, column := range columns {
		if !isValidColumn(q, column) {
			errorMessage := fmt.Sprintf(
				"invalid column: %s is not available on \"%s\" table.",
				column,
				GetTable(q.model),
			)

			raiden.Fatal(errorMessage)
		}
	}

	q.Columns = columns
	return &q
}

func GetColumnList(m interface{}) []string {
	r := reflect.TypeOf(m)

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

func isValidColumn(q Query, column string) bool {
	if column == "*" {
		return true
	}

	validSet := make(map[string]bool)

	for _, v := range GetColumnList(q.model) {
		validSet[v] = true
	}

	if validSet[column] {
		return true
	}

	return false
}
