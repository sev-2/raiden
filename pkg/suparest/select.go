package suparest

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/sev-2/raiden"
)

func (q Query) Select(columns ...string) (model *Query) {
	validSet := make(map[string]bool)

	for _, v := range GetColumnList(q.model) {
		validSet[v] = true
	}

	for _, v := range columns {
		if !validSet[v] {
			if strings.Contains(v, ":") {
				c := strings.Split(v, ":")
				isMatch, _ := regexp.MatchString(`^[a-zA-Z_][a-zA-Z0-9_]{1,59}`, c[0])
				if validSet[c[1]] && isMatch {
					continue
				}
			}

			raiden.Fatal(fmt.Sprintf("invalid column: %s is not available on %s table", v, GetTable(q.model)))
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
