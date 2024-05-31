package suparest

import (
	"encoding/json"
	"log"
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

	client := &fasthttp.Client{}

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	var cols []string
	t := reflect.TypeOf(p)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		cols = append(cols, field.Name)
	}

	req.SetRequestURI(url + "&" + strings.Join(cols, ","))

	serviceKey := getConfig().ServiceKey
	req.Header.SetMethod(fasthttp.MethodPatch)
	req.Header.SetContentType("application/json")
	req.Header.Set("apikey", serviceKey)
	req.Header.Set("Authorization", "Bearer "+serviceKey)
	req.Header.Set("Prefer", "return=representation")

	req.SetBody(jsonData)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	if err := client.Do(req, resp); err != nil {
		log.Fatalf("Error making GET request: %s\n", err)
	}

	body := resp.Body()

	return body, nil
}
