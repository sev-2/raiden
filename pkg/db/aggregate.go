package db

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/valyala/fasthttp"
)

type CountOptions struct {
	Count string
}

func (q *Query) Count(opts ...CountOptions) (int, error) {

	var countVal = "exact"

	for _, o := range opts {
		switch o.Count {
		case "exact":
			countVal = "exact"
		case "planned":
			countVal = "planned"
		case "estimated":
			countVal = "estimated"
		default:
			q.Errors = append(q.Errors, errors.New("unrecognized count options"))
		}
	}

	url := q.GetUrl()

	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	headers["Prefer"] = "count=" + countVal
	headers["Range-Unit"] = "items"

	_, resp, err := PostgrestRequest(q.Context, fasthttp.MethodHead, url, nil, headers)
	if err != nil {
		return 0, err
	}

	contentRange := resp.Header.Peek("Content-Range")
	parts := strings.Split(string(contentRange), "/")

	if len(parts) > 0 {
		return 0, errors.New("invalid range format")
	}

	count, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (q *Query) Sum(column string, alias string) *Query {
	err := validateColumn(*q, column)
	if err != nil {
		q.Errors = append(q.Errors, err)
		return nil
	}

	o := fmt.Sprintf("%s.%s", column, "sum()")

	if alias != "" {
		o = fmt.Sprintf("%s:%s", alias, o)
	}

	q.Columns = append(q.Columns, o)

	return q
}

func (q *Query) Avg(column string, alias string) *Query {
	err := validateColumn(*q, column)
	if err != nil {
		q.Errors = append(q.Errors, err)
		return nil
	}

	o := fmt.Sprintf("%s.%s", column, "avg()")

	if alias != "" {
		o = fmt.Sprintf("%s:%s", alias, o)
	}

	q.Columns = append(q.Columns, o)

	return q
}

func (q *Query) Min(column string, alias string) *Query {
	err := validateColumn(*q, column)
	if err != nil {
		q.Errors = append(q.Errors, err)
		return nil
	}

	o := fmt.Sprintf("%s.%s", column, "min()")

	if alias != "" {
		o = fmt.Sprintf("%s:%s", alias, o)
	}

	q.Columns = append(q.Columns, o)

	return q
}

func (q *Query) Max(column string, alias string) *Query {
	err := validateColumn(*q, column)
	if err != nil {
		q.Errors = append(q.Errors, err)
		return nil
	}

	o := fmt.Sprintf("%s.%s", column, "max()")

	if alias != "" {
		o = fmt.Sprintf("%s:%s", alias, o)
	}

	q.Columns = append(q.Columns, o)

	return q
}

func validateColumn(q Query, column string) error {
	table := GetTable(q.model)

	if !isColumnExist(q.model, column) {
		err := fmt.Errorf("invalid column: \"%s\" is not available on \"%s\" table", column, table)
		return err
	}

	if !isValidColumnName(column) {
		err := fmt.Errorf("invalid column: \"%s\" name is invalid", column)
		return err
	}

	if !isColumnExist(q.model, column) {
		err := fmt.Errorf("invalid alias column: \"%s\" is invalid or not available on \"%s\" table", column, table)
		return err
	}

	if !isValidColumnName(column) {
		err := fmt.Errorf("invalid alias column: \"%s\" name is invalid", column)
		return err
	}

	return nil
}
