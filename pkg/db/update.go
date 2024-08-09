package db

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/valyala/fasthttp"
)

func (q *Query) Update(p interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}

	url := q.GetUrl()

	var cols []string
	t := reflect.TypeOf(p)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		cols = append(cols, field.Name)
	}

	url = url + "&" + strings.Join(cols, ",")

	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	headers["Prefer"] = "return=representation"

	body, _, err := PostgrestRequest(q.Context, fasthttp.MethodPatch, url, jsonData, headers)
	if err != nil {
		return nil, err
	}

	return body, nil
}
