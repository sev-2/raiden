package db

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/resource"
)

func (q *Query) With(r string, columns map[string][]string, fkeys map[string]string) *Query {

	relations := strings.Split(r, ".")

	if len(relations) > 3 {
		raiden.Fatal("unsupported nested relations more than 3 levels")
	}

	regs := make(map[string]bool)

	for _, m := range resource.RegisteredModels {
		registeredModel := reflect.TypeOf(m).Elem().Name()
		regs[registeredModel] = true
	}

	for _, m := range relations {
		if !regs[m] {
			raiden.Fatal(fmt.Sprintf("invalid model name: %s", m))
		}
	}

	var selects []string

	for _, r := range reverseSortString(relations) {
		table := GetTable(findModel(resource.RegisteredModels, r))

		if keyExist(fkeys, r) {
			table = table + "!" + fkeys[r]
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

	q.Columns = append(q.Columns, selects...)

	return q
}
