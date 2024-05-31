package suparest

import (
	"log"

	"github.com/valyala/fasthttp"
)

func (q *Query) Delete() ([]byte, error) {
	url := q.GetUrl()

	client := &fasthttp.Client{}

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(url)

	serviceKey := getConfig().ServiceKey
	req.Header.SetMethod(fasthttp.MethodDelete)
	req.Header.SetContentType("application/json")
	req.Header.Set("apikey", serviceKey)
	req.Header.Set("Authorization", "Bearer "+serviceKey)
	req.Header.Set("Prefer", "return=representation")

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	if err := client.Do(req, resp); err != nil {
		log.Fatalf("Error making GET request: %s\n", err)
	}

	body := resp.Body()

	return body, nil
}

func (m ModelBase) ForceDelete() (model ModelBase) {
	return m
}
