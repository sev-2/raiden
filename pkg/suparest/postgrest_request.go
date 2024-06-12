package suparest

import (
	"fmt"
	"strings"

	"github.com/sev-2/raiden"
	"github.com/valyala/fasthttp"
)

func PostgrestRequest(ctx raiden.Context, method string, url string, payload []byte, headers map[string]string) ([]byte, error) {

	if !isAllowedMethod(method) {
		return nil, fmt.Errorf("method %s is not allowed", method)
	}

	client := &fasthttp.Client{}

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		baseUrl := getConfig().SupabasePublicUrl
		if strings.HasPrefix(url, "/") {
			url = fmt.Sprintf("%s/rest/v1%s", baseUrl, url)
		} else {
			url = fmt.Sprintf("%s/rest/v1/%s", baseUrl, url)
		}
	}

	req.SetRequestURI(url)
	req.Header.SetMethod(method)

	apikey := string(ctx.RequestContext().Request.Header.Peek("apikey"))
	if apikey != "" {
		req.Header.Set("apikey", apikey)
	}

	bearerToken := string(ctx.RequestContext().Request.Header.Peek("Authorization"))
	if bearerToken != "" && strings.HasPrefix(bearerToken, "Bearer ") {
		bearerToken = strings.TrimSpace(strings.TrimPrefix(bearerToken, "Bearer "))
		req.Header.Set("Authorization", bearerToken)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	if payload != nil {
		req.SetBody(payload)
	}

	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)

	if err := client.Do(req, res); err != nil {
		return nil, err
	}

	body := res.Body()

	return body, nil
}

func isAllowedMethod(method string) bool {
	methods := []string{
		fasthttp.MethodGet,
		fasthttp.MethodPost,
		fasthttp.MethodPut,
		fasthttp.MethodPatch,
		fasthttp.MethodDelete,
		fasthttp.MethodHead,
		fasthttp.MethodOptions,
	}

	for _, item := range methods {
		if strings.ToLower(item) == strings.ToLower(method) {
			return true
		}
	}
	return false
}
