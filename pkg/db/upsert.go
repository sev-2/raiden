package db

import (
	"encoding/json"

	"github.com/valyala/fasthttp"
)

type UpsertOptions = struct {
	OnConflict string
}

const (
	MergeDuplicates  = "merge-duplicates"
	IgnoreDuplicates = "ignore-duplicates"
)

func (q *Query) Upsert(payload []interface{}, opt UpsertOptions) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	url := q.GetUrl()

	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	headers["Prefer"] = "resolution=" + opt.OnConflict

	var a interface{}
	_, err := PostgrestRequestBind(q.Context, fasthttp.MethodPost, url, jsonData, headers, q.ByPass, &a)
	if err != nil {
		return err
	}

	return nil
}
