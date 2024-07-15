package db

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/sev-2/raiden"
)

func (q Query) Select(columns []string, aliases map[string]string) (model *Query) {

	table := GetTable(q.model)

	for _, column := range columns {
		if !isColumnExist(q.model, column) {
			err := fmt.Sprintf("invalid column: \"%s\" is not available on \"%s\" table.", column, table)
			raiden.Fatal(err)
		}

		if !isValidColumnName(column) {
			err := fmt.Sprintf("invalid column: \"%s\" name is invalid.", column)
			raiden.Fatal(err)
		}
	}

	for column, _ := range aliases {
		if !isColumnExist(q.model, column) {
			err := fmt.Sprintf("invalid alias column: \"%s\" is invalid or not available on \"%s\" table.", column, table)
			raiden.Fatal(err)
		}

		if !isValidColumnName(column) {
			err := fmt.Sprintf("invalid alias column: \"%s\" name is invalid.", column)
			raiden.Fatal(err)
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

func isColumnExist(model interface{}, column string) bool {
	if column == "*" {
		return true
	}

	validSet := make(map[string]bool)

	for _, v := range GetColumnList(model) {
		validSet[v] = true
	}

	return validSet[column]
}

func isValidColumnName(column string) bool {
	isAllowed, _ := regexp.MatchString(`^[a-zA-Z_][a-zA-Z0-9_]{1,59}`, column)

	return isAllowed
}
