package db

import (
	"fmt"
)

func (q *Query) OrderAsc(column string) *Query {
	if q.OrderList == nil {
		q.OrderList = &[]string{}
	}

	*q.OrderList = append(*q.OrderList, fmt.Sprintf("%s.asc", column))

	return q
}

func (q *Query) OrderDesc(column string) *Query {
	if q.OrderList == nil {
		q.OrderList = &[]string{}
	}

	*q.OrderList = append(*q.OrderList, fmt.Sprintf("%s.desc", column))

	return q
}
