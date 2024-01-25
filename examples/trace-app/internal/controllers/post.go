package controllers

import (
	"encoding/json"
	"fmt"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/tracer"
	"github.com/valyala/fasthttp"
)

type Post struct {
	Id     int    `json:"id"`
	UserId int    `json:"userId"`
	Title  string `json:"title"`
	Body   string `json:"body"`
}

type PostResponse []Post

// @type http-handler
// @route GET /posts
func PostController(ctx raiden.Context) raiden.Presenter {
	response, err := fetchPostsData(ctx)
	if err != nil {
		return ctx.SendJsonError(err)
	}

	return ctx.SendJson(response)
}

func fetchPostsData(ctx raiden.Context) (PostResponse, error) {
	var client fasthttp.Client
	var responseData PostResponse

	// Define the URL
	url := "https://jsonplaceholder.typicode.com/posts"

	// Create a new request
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	// Set the request method to GET
	req.Header.SetMethod(fasthttp.MethodGet)

	// Set the request URI
	req.SetRequestURI(url)

	// inject trace to request
	_, span := tracer.Inject(ctx.Context(), ctx.Tracer(), req)
	defer span.End()

	raiden.Infof("Added %s = %s", tracer.TraceIdHeaderKey, string(req.Header.Peek(tracer.TraceIdHeaderKey)))
	raiden.Infof("Added %s = %s", tracer.SpanIdHeaderKey, string(req.Header.Peek(tracer.SpanIdHeaderKey)))

	// Create a new response
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	if err := client.Do(req, resp); err != nil {
		err = fmt.Errorf("error making HTTP request: %v", err)
		span.RecordError(err)
		return responseData, err
	}

	if resp.StatusCode() == fasthttp.StatusOK {
		if err := json.Unmarshal(resp.Body(), &responseData); err != nil {
			err = fmt.Errorf("error unmarshalling response: %v", err)
			span.RecordError(err)
			return responseData, err
		}
	} else {
		err := fmt.Errorf("http request failed with status code: %d", resp.StatusCode())
		span.RecordError(err)
		return responseData, err
	}

	return responseData, nil
}
