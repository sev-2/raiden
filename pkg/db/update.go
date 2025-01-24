package db

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/valyala/fasthttp"
)

func (q *Query) Update(p interface{}, model interface{}) error {
	jsonData, err := json.Marshal(p)
	if err != nil {
		return err
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

	_, err0 := PostgrestRequest(q.Context, q.credential, fasthttp.MethodPatch, url, jsonData, headers, q.ByPass, model)
	if err0 != nil {
		return err0
	}

	return nil
}
