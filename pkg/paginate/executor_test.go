package paginate_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/client"
	"github.com/sev-2/raiden/pkg/mock"
	"github.com/sev-2/raiden/pkg/paginate"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

var mockData = []map[string]any{
	{"id": 1, "name": "test_1"},
	{"id": 2, "name": "test_2"},
	{"id": 3, "name": "test_3"},
}

var getMockCtx = func() *mock.MockContext {
	return &mock.MockContext{
		ConfigFn: func() *raiden.Config {
			return &raiden.Config{
				Mode:              raiden.BffMode,
				AnonKey:           "test_anon_key",
				ServiceKey:        "test_service_key",
				SupabasePublicUrl: "http://localhost:8002",
			}
		},
	}
}

var setMockRequest = func(interceptorFn func(r1 *fasthttp.Request, r2 *fasthttp.Response) error) func() {
	c := mock.MockClient{
		DoTimeoutFn: func(r1 *fasthttp.Request, r2 *fasthttp.Response, d time.Duration) error {
			if interceptorFn != nil {
				return interceptorFn(r1, r2)
			}
			return nil
		},
	}
	client.GetClientFn = func() client.Client { return &c }
	return func() {
		client.GetClientFn = client.DefaultGetClientFn
	}
}

func TestExecutor_BffOffset(t *testing.T) {
	// - Test normal case
	closeFn := setMockRequest(func(r1 *fasthttp.Request, r2 *fasthttp.Response) error {
		uri := r1.URI().String()
		assert.Equal(t, "http://localhost:8002/rest/v1/data?limit=10&offset=0", uri)
		dataByte, err := json.Marshal(mockData)
		if err != nil {
			return err
		}
		r2.SetBodyRaw(dataByte)
		r2.Header.Set("content-range", "0-24/3")
		return nil
	})
	defer closeFn()

	ctx := getMockCtx()
	paginator := paginate.NewFromContext(ctx, paginate.ExecuteOptions{
		Page:      1,
		Limit:     10,
		Type:      paginate.OffsetPagination,
		IsBypass:  true,
		WithCount: true,
	})

	result, err := paginator.Execute(context.Background(), "/data")
	assert.NoError(t, err)
	assert.Len(t, result.Data, 3)
	assert.Equal(t, 3, result.Count)

	// - Test paginate with filter
	closeFn2 := setMockRequest(func(r1 *fasthttp.Request, r2 *fasthttp.Response) error {
		uri := r1.URI().String()
		assert.Equal(t, "http://localhost:8002/rest/v1/data?select=*&limit=10&offset=0", uri)
		dataByte, err := json.Marshal(mockData)
		if err != nil {
			return err
		}
		r2.SetBodyRaw(dataByte)
		r2.Header.Set("content-range", "0-24/3")
		return nil
	})
	defer closeFn2()

	paginator2 := paginate.NewFromContext(ctx, paginate.ExecuteOptions{
		Page:      1,
		Limit:     10,
		Type:      paginate.OffsetPagination,
		IsBypass:  true,
		WithCount: true,
	})

	result2, err := paginator2.Execute(context.Background(), "/data?select=*")
	assert.NoError(t, err)
	assert.Len(t, result2.Data, 3)
	assert.Equal(t, 3, result2.Count)

	// - Test paginate page
	closeFn3 := setMockRequest(func(r1 *fasthttp.Request, r2 *fasthttp.Response) error {
		uri := r1.URI().String()
		assert.Equal(t, "http://localhost:8002/rest/v1/data?limit=10&offset=20", uri)
		dataByte, err := json.Marshal(mockData)
		if err != nil {
			return err
		}
		r2.SetBodyRaw(dataByte)
		r2.Header.Set("content-range", "0-24/3")
		return nil
	})
	defer closeFn3()

	paginator3 := paginate.NewFromContext(ctx, paginate.ExecuteOptions{
		Page:      3,
		Limit:     10,
		Type:      paginate.OffsetPagination,
		IsBypass:  true,
		WithCount: true,
	})

	result3, err := paginator3.Execute(context.Background(), "/data")
	assert.NoError(t, err)
	assert.Len(t, result3.Data, 3)
	assert.Equal(t, 3, result3.Count)
}

func TestExecutor_BffCursorNext(t *testing.T) {
	// - Test cursor first page
	closeFn := setMockRequest(func(r1 *fasthttp.Request, r2 *fasthttp.Response) error {
		uri := r1.URI().String()
		expectUri := fmt.Sprintf("http://localhost:8002/rest/v1/data?limit=%d", 4)
		assert.Equal(t, expectUri, uri)

		finalMockData := mockData
		finalMockData = append(finalMockData, map[string]any{"id": 4, "name": "test_4"})

		dataByte, err := json.Marshal(finalMockData)
		if err != nil {
			return err
		}

		r2.SetBodyRaw(dataByte)
		r2.Header.Set("content-range", "0-24/3")
		return nil
	})
	defer closeFn()

	ctx := getMockCtx()
	paginator := paginate.NewFromContext(ctx, paginate.ExecuteOptions{
		Limit:           3,
		Type:            paginate.CursorPagination,
		IsBypass:        true,
		WithCount:       true,
		CursorDirection: paginate.CursorPaginateDirectionNext,
	})

	result, err := paginator.Execute(context.Background(), "/data")

	rByte, _ := json.MarshalIndent(result, "*", " ")
	slog.Info(string(rByte))

	assert.NoError(t, err)
	assert.Len(t, result.Data, 3)
	assert.Equal(t, 3, result.Count)
	assert.Equal(t, nil, result.PrevCursor)
	assert.Equal(t, float64(3), result.NextCursor)

	// - Test cursor second page
	closeFn2 := setMockRequest(func(r1 *fasthttp.Request, r2 *fasthttp.Response) error {
		uri := r1.URI().String()
		expectUri := fmt.Sprintf("http://localhost:8002/rest/v1/data?select=*&id=gt.%d&limit=%d", 1, 4)
		if strings.Contains(uri, "lt.") {
			expectUri = "http://localhost:8002/rest/v1/data?select=*&id=lt.2&limit=1"
		}

		assert.Equal(t, expectUri, uri)
		data := []map[string]any{mockData[1], mockData[2]}
		if strings.Contains(uri, "lt.") {
			data = []map[string]any{mockData[0]}
		}
		dataByte, err := json.Marshal(data)
		if err != nil {
			return err
		}
		r2.SetBodyRaw(dataByte)
		r2.Header.Set("content-range", "0-24/2")
		return nil
	})
	defer closeFn2()

	paginator = paginate.NewFromContext(ctx, paginate.ExecuteOptions{
		Cursor:          1,
		Limit:           3,
		Type:            paginate.CursorPagination,
		IsBypass:        true,
		WithCount:       true,
		CursorDirection: paginate.CursorPaginateDirectionNext,
	})

	result, err = paginator.Execute(context.Background(), "/data?select=*")
	assert.NoError(t, err)
	assert.Len(t, result.Data, 2)
	assert.Equal(t, 2, result.Count)
	assert.Equal(t, float64(2), result.PrevCursor)
	assert.Nil(t, result.NextCursor)

	// - Test cursor with next cursor
	closeFn3 := setMockRequest(func(r1 *fasthttp.Request, r2 *fasthttp.Response) error {
		uri := r1.URI().String()
		expectUri := fmt.Sprintf("http://localhost:8002/rest/v1/data?select=*&limit=%d", 3)
		if strings.Contains(uri, "lt.") {
			expectUri = "http://localhost:8002/rest/v1/data?select=*&id=lt.1&limit=1"
		}

		assert.Equal(t, expectUri, uri)
		data := mockData
		if strings.Contains(uri, "lt.") {
			data = []map[string]any{mockData[0]}
		}
		dataByte, err := json.Marshal(data)
		if err != nil {
			return err
		}
		r2.SetBodyRaw(dataByte)
		r2.Header.Set("content-range", "0-24/2")
		return nil
	})
	defer closeFn3()

	paginator = paginate.NewFromContext(ctx, paginate.ExecuteOptions{
		Cursor:          nil,
		Limit:           2,
		Type:            paginate.CursorPagination,
		IsBypass:        true,
		WithCount:       true,
		CursorDirection: paginate.CursorPaginateDirectionNext,
	})

	result, err = paginator.Execute(context.Background(), "/data?select=*")
	assert.NoError(t, err)
	assert.Len(t, result.Data, 2)
	assert.Equal(t, 2, result.Count)
	assert.Equal(t, nil, result.PrevCursor)
	assert.Equal(t, float64(2), result.NextCursor)
}

func TestExecutor_BffCursorPrev(t *testing.T) {
	// - Test cursor first page
	token := "Bearer xxxxx"
	apiKey := "api_key_xxx"
	closeFn := setMockRequest(func(r1 *fasthttp.Request, r2 *fasthttp.Response) error {
		// assert uri
		uri := r1.URI().String()
		expectedUri := fmt.Sprintf("http://localhost:8002/rest/v1/data?limit=%d&order=id.desc", 4)
		if strings.Contains(uri, "limit=1.") {
			expectedUri = "http://localhost:8002/rest/v1/data?id=gt.1&limit=1&order=id.desc"
		}

		assert.Equal(t, expectedUri, uri)

		// assert token
		assert.Equal(t, token, string(r1.Header.Peek("Authorization")))
		assert.Equal(t, apiKey, string(r1.Header.Peek("apiKey")))

		finalMockData := []map[string]any{{"id": 4, "name": "test_4"}, mockData[2], mockData[1]}
		if strings.Contains(uri, "limit=1") {
			finalMockData = []map[string]any{{"id": 5, "name": "test_5"}}
		}

		dataByte, err := json.Marshal(finalMockData)
		if err != nil {
			return err
		}
		r2.SetBodyRaw(dataByte)
		r2.Header.Set("content-range", "0-24/3")
		return nil
	})
	defer closeFn()

	ctx := getMockCtx()

	requestCtx := &fasthttp.RequestCtx{
		Request:  fasthttp.Request{},
		Response: fasthttp.Response{},
	}
	requestCtx.Request.Header.Set("Authorization", token)
	requestCtx.Request.Header.Set("apikey", apiKey)
	ctx.RequestContextFn = func() *fasthttp.RequestCtx {
		return requestCtx
	}
	ctx.SendJsonFn = func(data any) error {
		return nil
	}

	paginator := paginate.NewFromContext(ctx, paginate.ExecuteOptions{
		Limit:           3,
		Type:            paginate.CursorPagination,
		WithCount:       true,
		CursorDirection: paginate.CursorPaginateDirectionPrev,
	})

	result, err := paginator.Execute(context.Background(), "/data")
	assert.NoError(t, err)
	assert.Len(t, result.Data, 3)
	assert.Equal(t, 3, result.Count)
	assert.Equal(t, nil, result.PrevCursor)
	assert.Equal(t, nil, result.NextCursor)

	// - Test cursor second page
	closeFn = setMockRequest(func(r1 *fasthttp.Request, r2 *fasthttp.Response) error {
		// assert uri
		uri := r1.URI().String()
		expectedUri := fmt.Sprintf("http://localhost:8002/rest/v1/data?%s=lt.%v&limit=%d&order=id.desc", paginate.DefaultOffsetColumn, 5, 4)
		if strings.Contains(uri, "limit=1") {
			expectedUri = "http://localhost:8002/rest/v1/data?id=gt.4&limit=1&order=id.desc"
		}

		assert.Equal(t, expectedUri, uri)

		// assert token
		assert.Equal(t, token, string(r1.Header.Peek("Authorization")))
		assert.Equal(t, apiKey, string(r1.Header.Peek("apiKey")))

		finalMockData := []map[string]any{{"id": 4, "name": "test_4"}, mockData[2], mockData[1], mockData[0]}
		if strings.Contains(uri, "limit=1") {
			finalMockData = []map[string]any{{"id": 5, "name": "test_5"}}
		}

		dataByte, err := json.Marshal(finalMockData)
		if err != nil {
			return err
		}
		r2.SetBodyRaw(dataByte)
		r2.Header.Set("content-range", "0-24/3")
		return nil
	})
	defer closeFn()

	paginator = paginate.NewFromContext(ctx, paginate.ExecuteOptions{
		Cursor:          5,
		Limit:           3,
		Type:            paginate.CursorPagination,
		WithCount:       true,
		CursorDirection: paginate.CursorPaginateDirectionPrev,
	})

	result, err = paginator.Execute(context.Background(), "/data")

	assert.NoError(t, err)
	assert.Len(t, result.Data, 3)
	assert.Equal(t, 3, result.Count)
	assert.Equal(t, float64(4), result.PrevCursor)
	assert.Equal(t, float64(2), result.NextCursor)

	// - Marshall test
	type mockDataStruct struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	marshallRs, err := paginate.MarshallResult[mockDataStruct](result)
	assert.NoError(t, err)
	assert.Len(t, marshallRs.Data, 3)
	assert.Equal(t, marshallRs.Count, 3)
	assert.Equal(t, float64(4), marshallRs.PrevCursor)
	assert.Equal(t, float64(2), marshallRs.NextCursor)

	// - Send Response Test
	err = paginate.SendResponse(ctx, result)
	assert.NoError(t, err)

	contentRange := requestCtx.Response.Header.Peek("content-range")
	assert.Equal(t, "0-24/3", string(contentRange))

	nextCursor := requestCtx.Response.Header.Peek("next-cursor")
	assert.Equal(t, "2", string(nextCursor))

	prevCursor := requestCtx.Response.Header.Peek("prev-cursor")
	assert.Equal(t, "4", string(prevCursor))
}

func TestExecutor_SvcCursor(t *testing.T) {
	// - Test normal case
	closeFn := setMockRequest(func(r1 *fasthttp.Request, r2 *fasthttp.Response) error {
		uri := r1.URI().String()
		assert.Equal(t, "http://localhost:8004/data?limit=10&offset=0", uri)

		assert.Equal(t, "Bearer jwt_token_xxxxx", string(r1.Header.Peek("Authorization")))

		dataByte, err := json.Marshal(mockData)
		if err != nil {
			return err
		}
		r2.SetBodyRaw(dataByte)
		r2.Header.Set("content-range", "0-24/3")
		return nil
	})
	defer closeFn()

	ctx := getMockCtx()
	ctx.ConfigFn = func() *raiden.Config {
		return &raiden.Config{
			PostgRestUrl: "http://localhost:8004",
			Mode:         raiden.SvcMode,
			JwtToken:     "jwt_token_xxxxx",
		}
	}

	paginator := paginate.NewFromContext(ctx, paginate.ExecuteOptions{
		Page:      1,
		Limit:     10,
		Type:      paginate.OffsetPagination,
		IsBypass:  true,
		WithCount: true,
	})

	result, err := paginator.Execute(context.Background(), "/data")
	assert.NoError(t, err)
	assert.Len(t, result.Data, 3)
	assert.Equal(t, 3, result.Count)
}

func TestExecutor_NewSvc(t *testing.T) {
	// - Test normal case
	closeFn := setMockRequest(func(r1 *fasthttp.Request, r2 *fasthttp.Response) error {
		uri := r1.URI().String()
		assert.Equal(t, "http://localhost:8004/data?limit=10&offset=0", uri)

		assert.Equal(t, "Bearer jwt_token_xxxxx", string(r1.Header.Peek("Authorization")))

		dataByte, err := json.Marshal(mockData)
		if err != nil {
			return err
		}
		r2.SetBodyRaw(dataByte)
		r2.Header.Set("content-range", "0-24/3")
		return nil
	})
	defer closeFn()

	cfg := &raiden.Config{
		PostgRestUrl: "http://localhost:8004",
		Mode:         raiden.SvcMode,
		JwtToken:     "jwt_token_xxxxx",
	}

	paginator := paginate.New(cfg, paginate.ExecuteOptions{
		Page:      1,
		Limit:     10,
		Type:      paginate.OffsetPagination,
		IsBypass:  true,
		WithCount: true,
	})

	result, err := paginator.Execute(context.Background(), "/data")
	assert.NoError(t, err)
	assert.Len(t, result.Data, 3)
	assert.Equal(t, 3, result.Count)
}

func TestExecutor_NewBffCursorNext(t *testing.T) {
	closeFn := setMockRequest(func(r1 *fasthttp.Request, r2 *fasthttp.Response) error {
		uri := r1.URI().String()
		expectUri := fmt.Sprintf("http://localhost:8002/rest/v1/data?select=*&limit=%d", 4)
		assert.Equal(t, expectUri, uri)

		finalMockData := mockData
		finalMockData = append(finalMockData, map[string]any{"id": 4, "name": "test_4"})

		dataByte, err := json.Marshal(finalMockData)
		if err != nil {
			return err
		}
		r2.SetBodyRaw(dataByte)
		r2.Header.Set("content-range", "0-24/3")
		return nil
	})
	defer closeFn()

	ctx := getMockCtx()
	cfg := ctx.Config()
	cfg.SupabasePublicUrl = "http://localhost:8002/"
	paginator := paginate.New(cfg, paginate.ExecuteOptions{
		Limit:           3,
		Type:            paginate.CursorPagination,
		IsBypass:        true,
		WithCount:       true,
		CursorDirection: paginate.CursorPaginateDirectionNext,
	})

	result, err := paginator.Execute(context.Background(), "/data?select=*")
	assert.NoError(t, err)
	assert.Len(t, result.Data, 3)
	assert.Equal(t, 3, result.Count)
	assert.Equal(t, nil, result.PrevCursor)
	assert.Equal(t, float64(3), result.NextCursor)
}

func TestExecutor_NewBffCursorPrev(t *testing.T) {
	closeFn := setMockRequest(func(r1 *fasthttp.Request, r2 *fasthttp.Response) error {
		uri := r1.URI().String()
		expectedUri := fmt.Sprintf("http://localhost:8002/rest/v1/data?select=*&limit=%d&order=id.desc", 4)
		assert.Equal(t, expectedUri, uri)
		return fasthttp.ErrConnectionClosed
	})
	defer closeFn()

	ctx := getMockCtx()
	cfg := ctx.Config()
	cfg.SupabasePublicUrl = "http://localhost:8002/"
	paginator := paginate.New(cfg, paginate.ExecuteOptions{
		Limit:           3,
		Type:            paginate.CursorPagination,
		IsBypass:        true,
		WithCount:       true,
		CursorDirection: paginate.CursorPaginateDirectionPrev,
	})

	_, err := paginator.Execute(context.Background(), "/data?select=*")
	assert.Error(t, err)
}
