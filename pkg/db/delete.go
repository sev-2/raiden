package db

import (
	"github.com/valyala/fasthttp"
)

func (q *Query) Delete() error {
	url := q.GetUrl()

	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	headers["Prefer"] = "return=representation"

	var a interface{}
	_, err := PostgrestRequest(q.Context, q.credential, fasthttp.MethodDelete, url, nil, headers, q.ByPass, &a)
	if err != nil {
		return err
	}

	return nil
}
