package db

import (
	"errors"
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

	count, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, err
	}

	return count, nil
}
