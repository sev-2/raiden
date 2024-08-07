package db

import (
	"encoding/json"

	"github.com/valyala/fasthttp"
)

func (q *Query) Insert(payload interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	url := q.GetUrl()

	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	headers["Prefer"] = "return=representation"

	body, _, err := PostgrestRequest(q.Context, fasthttp.MethodPost, url, jsonData, headers)
	if err != nil {
		return nil, err
	}

	return body, nil
}
