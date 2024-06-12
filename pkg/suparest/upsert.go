package suparest

import (
	"encoding/json"
	"log"

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

	client := &fasthttp.Client{}

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(url)

	serviceKey := getConfig().ServiceKey
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetContentType("application/json")
	req.Header.Set("apikey", serviceKey)
	req.Header.Set("Authorization", "Bearer "+serviceKey)
	req.Header.Set("Prefer", "resolution="+opt.OnConflict)

	req.SetBody(jsonData)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	if err := client.Do(req, resp); err != nil {
		log.Fatalf("Error making GET request: %s\n", err)
	}

	body := resp.Body()

	return body, nil
}
