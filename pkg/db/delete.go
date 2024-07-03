package db

import (
	"github.com/valyala/fasthttp"
)

func (q *Query) Delete() ([]byte, error) {
	url := q.GetUrl()

	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	headers["Prefer"] = "return=representation"

	resp, _, err := PostgrestRequest(q.Context, fasthttp.MethodDelete, url, nil, headers)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (m ModelBase) ForceDelete() (model ModelBase) {
	return m
}
