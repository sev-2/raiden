package suparest

import (
	"log"
	"strconv"
	"strings"

	"github.com/sev-2/raiden"
	"github.com/valyala/fasthttp"
)

type CountOptions struct {
	Count string
}

func (q *Query) Count(opts ...CountOptions) (int, error) {

	var countVal = "exact"

	for _, o := range opts {
		switch o.Count {
		case "exact":
			countVal = "exact"
		case "planned":
			countVal = "planned"
		case "estimated":
			countVal = "estimated"
		default:
			raiden.Fatal("Unrecognized count options.")
		}
	}

	url := q.GetUrl()

	client := &fasthttp.Client{}

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(url)

	serviceKey := getConfig().ServiceKey
	req.Header.SetMethod(fasthttp.MethodHead)
	req.Header.Set("apikey", serviceKey)
	req.Header.Set("Authorization", "Bearer "+serviceKey)
	req.Header.Set("Prefer", "count="+countVal)
	req.Header.Set("Range-Unit", "items")

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	if err := client.Do(req, resp); err != nil {
		log.Fatalf("Error making GET request: %s\n", err)
	}

	contentRange := resp.Header.Peek("Content-Range")
	parts := strings.Split(string(contentRange), "/")

	count, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, err
	}
	log.Println(string(contentRange), count)

	return count, nil
}
