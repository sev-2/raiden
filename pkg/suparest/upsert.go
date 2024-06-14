package suparest

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

func (q *Query) Upsert(payload []interface{}, opt UpsertOptions) ([]byte, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	url := q.GetUrl()

	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	headers["Prefer"] = "resolution=" + opt.OnConflict

	resp, _, err := PostgrestRequest(q.Context, fasthttp.MethodPost, url, jsonData, headers)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
