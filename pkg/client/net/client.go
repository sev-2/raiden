package net

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/sev-2/raiden/pkg/logger"
)

type Client interface {
	Do(req *http.Request) (*http.Response, error)
}

type (
	RequestInterceptor  func(req *http.Request) error
	ResponseInterceptor func(resp *http.Response) error
)

type ReqError struct {
	Message string
	Body    []byte
}

func (s ReqError) Error() string {
	return s.Message
}

var (
	headerContentTypeJson = "application/json"
	DefaultTimeout        = time.Duration(20000) * time.Millisecond
	Logger                = logger.HcLog().Named("supabase.client.net")
)

type DefaultResponse struct {
	Message string `json:"message"`
}

var GetClient func() Client = func() Client {
	return &http.Client{
		Timeout: DefaultTimeout,
	}
}

func SendRequest(method string, url string, body []byte, timeout time.Duration, reqInterceptor RequestInterceptor, resInterceptor ResponseInterceptor) (rawBody []byte, err error) {
	reqTimeout := DefaultTimeout
	if timeout != 0 {
		reqTimeout = timeout
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", headerContentTypeJson)

	// perform request interceptor when exist
	if reqInterceptor != nil {
		if err = reqInterceptor(req); err != nil {
			return nil, err
		}
	}

	Logger.Trace("net.request", "url", url)
	Logger.Trace("net.request", "timeout", reqTimeout)
	Logger.Trace("net.request", "body", string(body))

	resp, err := GetClient().Do(req)
	if err != nil {
		errName, known := ExtractResponseErr(err)
		if known {
			err = fmt.Errorf("conn error: %v", errName)
		} else {
			err = fmt.Errorf("conn failure: %v %v", errName, err)
		}
		Logger.Trace("net.request", "err-type", reflect.TypeOf(err).String(), "err-msg", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	if !strings.HasPrefix(strconv.Itoa(statusCode), "2") {
		err = fmt.Errorf("invalid HTTP response code: %d", statusCode)
		if resp.Body != nil {
			body, _ := io.ReadAll(resp.Body)
			Logger.Error(string(body))
			sendErr := ReqError{
				Message: err.Error(),
				Body:    body,
			}
			return nil, sendErr
		}
		return nil, err
	}

	if resInterceptor != nil {
		if err := resInterceptor(resp); err != nil {
			return body, err
		}
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		Logger.Error("net.response", "err-type", reflect.TypeOf(err).String(), "err-msg", err.Error())
		return nil, err
	}

	Logger.Trace("net.response", "body", string(body))
	return body, nil
}

func Post[T any](url string, rawBody []byte, timeout time.Duration, reqInterceptor RequestInterceptor, resInterceptor ResponseInterceptor) (res T, err error) {
	byteData, err := SendRequest(http.MethodPost, url, rawBody, timeout, reqInterceptor, resInterceptor)
	if err != nil {
		return res, err
	}
	return res, json.Unmarshal(byteData, &res)
}

func Get[T any](url string, timeout time.Duration, reqInterceptor RequestInterceptor, resInterceptor ResponseInterceptor) (res T, err error) {
	byteData, err := SendRequest(http.MethodGet, url, nil, timeout, reqInterceptor, resInterceptor)
	if err != nil {
		return res, err
	}
	return res, json.Unmarshal(byteData, &res)
}

func Patch[T any](url string, rawBody []byte, timeout time.Duration, reqInterceptor RequestInterceptor, resInterceptor ResponseInterceptor) (res T, err error) {
	byteData, err := SendRequest(http.MethodPatch, url, rawBody, timeout, reqInterceptor, resInterceptor)
	if err != nil {
		return res, err
	}
	return res, json.Unmarshal(byteData, &res)
}

func Put[T any](url string, rawBody []byte, timeout time.Duration, reqInterceptor RequestInterceptor, resInterceptor ResponseInterceptor) (res T, err error) {
	byteData, err := SendRequest(http.MethodPut, url, rawBody, timeout, reqInterceptor, resInterceptor)
	if err != nil {
		return res, err
	}
	return res, json.Unmarshal(byteData, &res)
}

func Delete[T any](url string, rawBody []byte, timeout time.Duration, reqInterceptor RequestInterceptor, resInterceptor ResponseInterceptor) (res T, err error) {
	byteData, err := SendRequest(http.MethodDelete, url, rawBody, timeout, reqInterceptor, resInterceptor)
	if err != nil {
		return res, err
	}
	return res, json.Unmarshal(byteData, &res)
}

func ExtractResponseErr(err error) (string, bool) {
	var (
		errName string
		known   = true
	)

	switch {
	case errors.Is(err, http.ErrHandlerTimeout):
		errName = "timeout"
	case errors.Is(err, http.ErrServerClosed):
		errName = "conn_close"
	default:
		known = false
	}

	return errName, known
}
