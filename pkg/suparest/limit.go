package suparest

func (q *Query) Limit(value int) *Query {
	q.LimitValue = value
	return q
}

func (q *Query) Offset(value int) *Query {
	q.OffsetValue = value
	return q
}
