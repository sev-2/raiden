package utils

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/fatih/color"
	"github.com/valyala/fasthttp"
)

var (
	headerContentTypeJson = []byte("application/json")
	httpClient            *fasthttp.Client
)

type SendRequestError struct {
	Message string
	Body    []byte
}

func (s SendRequestError) Error() string {
	return s.Message
}

func GetColoredHttpMethod(httpMethod string) string {
	printFunc := color.New(color.FgBlack, color.BgHiWhite).SprintfFunc()
	switch httpMethod {
	case http.MethodGet:
		printFunc = color.New(color.FgWhite, color.BgGreen).SprintfFunc()
	case http.MethodPost:
		printFunc = color.New(color.FgWhite, color.BgHiYellow).SprintfFunc()
	case http.MethodPatch:
		printFunc = color.New(color.FgWhite, color.BgHiBlue).SprintfFunc()
	case http.MethodPut:
		printFunc = color.New(color.FgWhite, color.BgBlue).SprintfFunc()
	case http.MethodDelete:
		printFunc = color.New(color.FgWhite, color.BgHiRed).SprintfFunc()
	}

	return printFunc(" %s ", httpMethod)
}

func getHttpClient() *fasthttp.Client {
	if httpClient == nil {
		readTimeout, _ := time.ParseDuration("1s")
		writeTimeout, _ := time.ParseDuration("2s")
		maxIdleConnDuration, _ := time.ParseDuration("1h")
		httpClient = &fasthttp.Client{
			ReadTimeout:                   readTimeout,
			WriteTimeout:                  writeTimeout,
			MaxIdleConnDuration:           maxIdleConnDuration,
			NoDefaultUserAgentHeader:      true, // Don't send: User-Agent: fasthttp
			DisableHeaderNamesNormalizing: true, // If you set the case on your headers correctly you can enable this
			DisablePathNormalizing:        true,
			// increase DNS cache time to an hour instead of default minute
			Dial: (&fasthttp.TCPDialer{
				Concurrency:      4096,
				DNSCacheDuration: time.Hour,
			}).Dial,
		}
	}
	return httpClient
}

func SendRequest(httpMethod string, url string, body []byte, reqInterceptor func(req *fasthttp.Request)) ([]byte, error) {

	// per-request timeout
	reqTimeout := time.Duration(1000) * time.Millisecond

	req := fasthttp.AcquireRequest()
	req.SetRequestURI(url)
	req.Header.SetMethod(httpMethod)
	req.Header.SetContentTypeBytes(headerContentTypeJson)
	req.SetBodyRaw(body)

	if reqInterceptor != nil {
		reqInterceptor(req)
	}

	// Create a new response
	resp := fasthttp.AcquireResponse()
	err := getHttpClient().DoTimeout(req, resp, reqTimeout)

	// release request and response
	fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	if err != nil {
		errName, known := httpConnError(err)
		if known {
			err = fmt.Errorf("conn error: %v", errName)
		} else {
			err = fmt.Errorf("conn failure: %v %v", errName, err)
		}
		return nil, err
	}

	statusCode := resp.StatusCode()

	if statusCode != http.StatusOK {
		err = fmt.Errorf("invalid HTTP response code: %d", statusCode)
		if resp.Body() != nil && len(resp.Body()) > 0 {
			sendErr := SendRequestError{
				Message: err.Error(),
				Body:    resp.Body(),
			}
			return nil, sendErr
		}
		return nil, err
	}

	return resp.Body(), nil
}

func httpConnError(err error) (string, bool) {
	var (
		errName string
		known   = true
	)

	switch {
	case errors.Is(err, fasthttp.ErrTimeout):
		errName = "timeout"
	case errors.Is(err, fasthttp.ErrNoFreeConns):
		errName = "conn_limit"
	case errors.Is(err, fasthttp.ErrConnectionClosed):
		errName = "conn_close"
	case reflect.TypeOf(err).String() == "*net.OpError":
		errName = "timeout"
	default:
		known = false
	}

	return errName, known
}
