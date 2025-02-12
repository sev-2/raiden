package db

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"strings"

	"github.com/sev-2/raiden"
	"github.com/valyala/fasthttp"
)

type ErrorResponse struct {
	Message string `json:"message"`
	Code    string `json:"code"`
	Hint    string `json:"hint"`
	Details string `json:"details"`
}

func PostgrestRequest(ctx raiden.Context, credential Credential, method string, url string, payload []byte, headers map[string]string, bypass bool, result interface{}) (*fasthttp.Response, error) {
	callbackErr := func(errCode int, data []byte) error {
		var errorResponse ErrorResponse
		if err := json.Unmarshal(data, &errorResponse); err == nil && errorResponse.Message != "" && errorResponse.Code != "" {
			return errors.New(errorResponse.Message)
		}
		return nil
	}

	if ctx != nil {
		return PostgrestRequestBind(ctx, method, url, payload, headers, bypass, result, callbackErr)
	}
	return PostgrestRequestBindCredential(credential, method, url, payload, headers, bypass, result, callbackErr)
}

func PostgrestRequestBind(ctx raiden.Context, method string, url string, payload []byte, headers map[string]string, bypass bool, result interface{}, callbackErr func(code int, data []byte) error) (*fasthttp.Response, error) {

	if !isAllowedMethod(method) {
		return nil, fmt.Errorf("method %s is not allowed", method)
	}

	client := &fasthttp.Client{}

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		var baseUrl string
		if flag.Lookup("test.v") != nil {
			baseUrl = "https://api.supabase.co"
		} else {
			baseUrl = getConfig().SupabasePublicUrl

			if getConfig().Mode == raiden.SvcMode {
				baseUrl = getConfig().PostgRestUrl
			}
		}

		if getConfig() != nil && getConfig().Mode == raiden.SvcMode {
			if strings.HasPrefix(url, "/") {
				url = fmt.Sprintf("%s%s", baseUrl, url)
			} else {
				url = fmt.Sprintf("%s/%s", baseUrl, url)
			}
		} else {
			if strings.HasPrefix(url, "/") {
				url = fmt.Sprintf("%s/rest/v1%s", baseUrl, url)
			} else {
				url = fmt.Sprintf("%s/rest/v1/%s", baseUrl, url)
			}
		}

	}

	req.SetRequestURI(url)
	req.Header.SetMethod(method)

	if bypass {
		if flag.Lookup("test.v") == nil {
			req.Header.Set("apikey", getConfig().ServiceKey)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", getConfig().ServiceKey))
		}

		if getConfig() != nil && getConfig().Mode == raiden.SvcMode && getConfig().JwtToken != "" {
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", getConfig().JwtToken))
		}
	} else {
		apikey := string(ctx.RequestContext().Request.Header.Peek("apikey"))
		if apikey != "" {
			req.Header.Set("apikey", apikey)
		}
		bearerToken := string(ctx.RequestContext().Request.Header.Peek("Authorization"))
		if bearerToken != "" && strings.HasPrefix(bearerToken, "Bearer ") {
			req.Header.Set("Authorization", bearerToken)
		}
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
		return res, err
	}

	body := res.Body()
	if callbackErr != nil {
		if err := callbackErr(res.StatusCode(), body); err != nil {
			return res, err
		}
	}

	if result != nil {
		if err := json.Unmarshal(body, result); err != nil {
			return res, fmt.Errorf("failed to unmarshal response body: %w", err)
		}
	}

	return res, nil
}

func PostgrestRequestBindCredential(credential Credential, method string, url string, payload []byte, headers map[string]string, bypass bool, result interface{}, callbackErr func(code int, data []byte) error) (*fasthttp.Response, error) {
	if !isAllowedMethod(method) {
		return nil, fmt.Errorf("method %s is not allowed", method)
	}

	client := &fasthttp.Client{}

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		var baseUrl string
		if flag.Lookup("test.v") != nil {
			baseUrl = "https://api.supabase.co"
		} else {
			baseUrl = getConfig().SupabasePublicUrl

			if getConfig().Mode == raiden.SvcMode {
				baseUrl = getConfig().PostgRestUrl
			}
		}

		if getConfig() != nil && getConfig().Mode == raiden.SvcMode {
			if strings.HasPrefix(url, "/") {
				url = fmt.Sprintf("%s%s", baseUrl, url)
			} else {
				url = fmt.Sprintf("%s/%s", baseUrl, url)
			}
		} else {
			if strings.HasPrefix(url, "/") {
				url = fmt.Sprintf("%s/rest/v1%s", baseUrl, url)
			} else {
				url = fmt.Sprintf("%s/rest/v1/%s", baseUrl, url)
			}
		}

	}

	req.SetRequestURI(url)
	req.Header.SetMethod(method)

	if bypass {
		if flag.Lookup("test.v") == nil {
			req.Header.Set("apikey", getConfig().ServiceKey)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", getConfig().ServiceKey))
		}

		if getConfig() != nil && getConfig().Mode == raiden.SvcMode && getConfig().JwtToken != "" {
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", getConfig().JwtToken))
		}
	} else {

		if credential.ApiKey != "" {
			req.Header.Set("apikey", credential.ApiKey)
		}

		if credential.Token != "" {
			if strings.HasPrefix(credential.Token, "Bearer ") {
				req.Header.Set("Authorization", credential.Token)
			} else {
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", credential.Token))
			}
		}
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
		return res, err
	}

	body := res.Body()

	if callbackErr != nil {
		if err := callbackErr(res.StatusCode(), body); err != nil {
			return res, err
		}
	}

	if result != nil {
		if err := json.Unmarshal(body, result); err != nil {
			return res, fmt.Errorf("failed to unmarshal response body: %w", err)
		}
	}

	return res, nil
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
		if strings.EqualFold(item, method) {
			return true
		}
	}
	return false
}
