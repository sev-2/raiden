package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/sev-2/raiden/pkg/logger"
	"github.com/valyala/fasthttp"
)

var (
	headerContentTypeJson = []byte("application/json")
	httpClientInstance    Client
	DefaultTimeout        = time.Duration(20000) * time.Millisecond
	Logger                = logger.HcLog().Named("supabase.client")
)

type Client interface {
	Do(*fasthttp.Request, *fasthttp.Response) error
	DoTimeout(*fasthttp.Request, *fasthttp.Response, time.Duration) error
}

type (
	RequestInterceptor  func(req *fasthttp.Request) error
	ResponseInterceptor func(resp *fasthttp.Response) error
)

type ReqError struct {
	Message string
	Body    []byte
}

func (s ReqError) Error() string {
	return s.Message
}

type DefaultResponse struct {
	Message string `json:"message"`
}

func getClient() Client {
	maxIdleConnDuration := time.Hour * 1
	return &fasthttp.Client{
		ReadTimeout:                   DefaultTimeout,
		WriteTimeout:                  DefaultTimeout,
		MaxIdleConnDuration:           maxIdleConnDuration,
		NoDefaultUserAgentHeader:      true,
		DisableHeaderNamesNormalizing: true,
		DisablePathNormalizing:        true,
		MaxConnWaitTimeout:            DefaultTimeout,
		Dial: (&fasthttp.TCPDialer{
			Concurrency:      4096,
			DNSCacheDuration: time.Hour,
		}).Dial,
	}
}

func SetClient(c Client) {
	httpClientInstance = c
}

func SendRequest(method string, url string, body []byte, timeout time.Duration, reqInterceptor RequestInterceptor, resInterceptor ResponseInterceptor) (rawBody []byte, err error) {
	reqTimeout := DefaultTimeout
	if timeout != 0 {
		reqTimeout = timeout
	}

	req, resp := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()

	req.SetRequestURI(url)
	req.Header.SetMethod(method)
	req.Header.SetContentTypeBytes(headerContentTypeJson)

	if len(body) > 0 {
		req.SetBodyRaw(body)
	}

	// perform request interceptor when exist
	if reqInterceptor != nil {
		if err = reqInterceptor(req); err != nil {
			return rawBody, err
		}
	}

	Logger.Trace("request", "timeout", reqTimeout)
	Logger.Trace("request", "headers", string(req.Header.Header()))
	Logger.Trace("request", "body", string(req.Body()))

	err = getClient().DoTimeout(req, resp, reqTimeout)
	if err != nil {
		logger.HcLog().Error("[SendRequest][DoTimeout]", "err-type", reflect.TypeOf(err).String(), "err-msg", err.Error())
		return
	}

	fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	if err != nil {
		errName, known := extractResponseErr(err)
		if known {
			err = fmt.Errorf("conn error: %v", errName)
		} else {
			err = fmt.Errorf("conn failure: %v %v", errName, err)
		}
		return nil, err
	}

	statusCode := resp.StatusCode()

	if !strings.HasPrefix(strconv.Itoa(statusCode), "2") {
		err = fmt.Errorf("invalid HTTP response code: %d", statusCode)
		if resp.Body() != nil && len(resp.Body()) > 0 {
			Logger.Error(string(resp.Body()))
			sendErr := ReqError{
				Message: err.Error(),
				Body:    resp.Body(),
			}
			return nil, sendErr
		}
		return nil, err
	}

	if resInterceptor != nil {
		if err := resInterceptor(resp); err != nil {
			return rawBody, err
		}
	}

	return resp.Body(), nil
}

func Post[T any](url string, rawBody []byte, timeout time.Duration, reqInterceptor RequestInterceptor, resInterceptor ResponseInterceptor) (res T, err error) {
	byteData, err := SendRequest(fasthttp.MethodPost, url, rawBody, timeout, reqInterceptor, resInterceptor)
	if err != nil {
		return res, err
	}
	return res, json.Unmarshal(byteData, &res)
}

func Get[T any](url string, timeout time.Duration, reqInterceptor RequestInterceptor, resInterceptor ResponseInterceptor) (res T, err error) {
	byteData, err := SendRequest(fasthttp.MethodGet, url, nil, timeout, reqInterceptor, resInterceptor)
	if err != nil {
		return res, err
	}
	return res, json.Unmarshal(byteData, &res)
}

func Patch[T any](url string, rawBody []byte, timeout time.Duration, reqInterceptor RequestInterceptor, resInterceptor ResponseInterceptor) (res T, err error) {
	byteData, err := SendRequest(fasthttp.MethodPatch, url, rawBody, timeout, reqInterceptor, resInterceptor)
	if err != nil {
		return res, err
	}
	return res, json.Unmarshal(byteData, &res)
}

func Put[T any](url string, rawBody []byte, timeout time.Duration, reqInterceptor RequestInterceptor, resInterceptor ResponseInterceptor) (res T, err error) {
	byteData, err := SendRequest(fasthttp.MethodPut, url, rawBody, timeout, reqInterceptor, resInterceptor)
	if err != nil {
		return res, err
	}
	return res, json.Unmarshal(byteData, &res)
}

func Delete[T any](url string, rawBody []byte, timeout time.Duration, reqInterceptor RequestInterceptor, resInterceptor ResponseInterceptor) (res T, err error) {
	byteData, err := SendRequest(fasthttp.MethodDelete, url, rawBody, timeout, reqInterceptor, resInterceptor)
	if err != nil {
		return res, err
	}
	return res, json.Unmarshal(byteData, &res)
}

func extractResponseErr(err error) (string, bool) {
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
		logger.HcLog().Error("[extractResponseErr][timeout]", "msg-err", reflect.TypeOf(err).String())
		errName = "timeout"
	default:
		known = false
	}

	return errName, known
}
