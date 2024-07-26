package db

import (
	"fmt"
	"strings"

	"github.com/sev-2/raiden"
)

func (q *Query) With(r string, columns map[string][]string) *Query {

	relations := strings.Split(r, ".")

	if len(relations) > 3 {
		raiden.Fatal("unsupported nested relations more than 3 levels")
	}

	for _, m := range relations {
		if findModel(m) == nil {
			raiden.Fatal(fmt.Sprintf("invalid model name: %s", m))
		}
	}

	var selects []string

	for _, r := range reverseSortString(relations) {
		model := findModel(r)
		table := GetTable(model)

		for k := range columns {
			if strings.Contains(k, "!") {
				split := strings.Split(k, "!")
				m := findModel(split[0])
				if !isForeignKeyExist(m, split[1]) {
					err := fmt.Sprintf("invalid foreign key: \"%s\" key is not exist.", split[1])
					raiden.Fatal(err)
				} else {
					table = fmt.Sprintf("%s!%s", table, split[1])
				}
			}
		}

		// Columns validations
		for _, c := range columns[r] {

			var column = c

			if strings.Contains(c, ":") {
				split := strings.Split(c, ":")
				alias := split[0]
				column = split[1]
				if !isValidColumnName(alias) {
					err := fmt.Sprintf("invalid alias column name: \"%s\" name is invalid.", alias)
					raiden.Fatal(err)
				}
			}

			if !isColumnExist(model, column) {
				err := fmt.Sprintf("invalid column: \"%s\" is not available on \"%s\" table.", column, table)
				raiden.Fatal(err)
			}

			if !isValidColumnName(column) {
				err := fmt.Sprintf("invalid column: \"%s\" name is invalid.", column)
				raiden.Fatal(err)
			}
		}

		cols := strings.Join(columns[r], ",")

		if len(cols) == 0 {
			cols = "*"
		}

		if len(selects) > 0 {
			lastQuery := selects[len(selects)-1]
			selects[len(selects)-1] = fmt.Sprintf("%s(%s,%s)", table, cols, lastQuery)
		} else {
			selects = append(selects, fmt.Sprintf("%s(%s)", table, cols))
		}
	}

	q.Relations = append(q.Relations, selects...)

	return q
}
