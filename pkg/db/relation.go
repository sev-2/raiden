package db

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/resource"
)

func (q *Query) With(r string, columns map[string][]string) *Query {

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
		tbl, _ := parseFkey(m)

		if !regs[tbl] {
			raiden.Fatal(fmt.Sprintf("invalid model name: %s", m))
		}
	}

	var selects []string

	for _, r := range reverseSortString(relations) {
		tbl, fkey := parseFkey(r)

		table := GetTable(findModel(resource.RegisteredModels, tbl))

		if fkey != "" {
			table = table + "!" + fkey
		}

		cols := strings.Join(columns[tbl], ",")

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

func findModel(models []interface{}, targetName string) interface{} {
	for _, m := range models {
		if reflect.TypeOf(m).Elem().Name() == targetName {
			return m
		}
	}

	return nil
}

func reverseSortString(n []string) []string {
	sort.Slice(n, func(i, j int) bool {
		return n[i] < n[j]
	})

	return n
}

func parseFkey(s string) (string, string) {
	t := strings.Split(s, "!")
	if len(t) > 1 {
		return t[0], t[1]
	}
	return t[0], ""
}
