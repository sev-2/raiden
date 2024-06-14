package suparest

func (q Query) Select(c ...string) (model *Query) {
	q.Columns = c
	return &q
}
