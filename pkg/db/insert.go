package db

import (
	"encoding/json"

	"github.com/valyala/fasthttp"
)

func (q *Query) Insert(payload interface{}, model interface{}) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := q.GetUrl()

	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	headers["Prefer"] = "return=representation"

	_, err0 := PostgrestRequestBind(q.Context, fasthttp.MethodPost, url, jsonData, headers, q.ByPass, model)
	if err0 != nil {
		return err0
	}

	return nil
}
