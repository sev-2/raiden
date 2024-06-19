package suparest

import (
	"fmt"
	"strings"
)

func (q *Query) With(m interface{}, cols []string) *Query {
	table := GetTable(m)

	columns := strings.Join(cols, ",")

	q.Columns = append(
		q.Columns,
		fmt.Sprintf("%s(%s)", table, columns),
	)

	return q
}
