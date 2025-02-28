package paginate

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/sev-2/raiden/pkg/client"
	"github.com/valyala/fasthttp"
)

type SupabaseDriver struct {
	baseUrl string
	apiKey  string
	token   string
}

func (s *SupabaseDriver) Paginate(ctx context.Context, statement string, page, limit int, withCount bool) ([]Item, int, error) {
	offset := (page - 1) * limit

	paginateStatement := fmt.Sprintf("limit=%d&offset=%d", limit, offset)
	if len(strings.Split(statement, "?")) == 2 {
		statement = fmt.Sprintf("%s&%s", statement, paginateStatement)
	} else {
		statement = fmt.Sprintf("%s?%s", statement, paginateStatement)
	}
	return s.request(statement, withCount)
}

func (s *SupabaseDriver) CursorPaginateNext(ctx context.Context, statement string, cursorRefColumn string, cursor any, limit int, withCount bool) ([]Item, int, any, any, error) {
	paginateStatement := fmt.Sprintf("limit=%d", limit)

	if cursor != nil {
		if reflect.TypeOf(cursor).Kind() == reflect.String && len(cursor.(string)) > 0 {
			paginateStatement = fmt.Sprintf("%s=gt.%v&limit=%d", cursorRefColumn, cursor, limit)
		}
		if reflect.TypeOf(cursor).Kind() == reflect.Int && cursor.(int) > 0 {
			paginateStatement = fmt.Sprintf("%s=gt.%v&limit=%d", cursorRefColumn, cursor, limit)
		}
	}

	if len(strings.Split(statement, "?")) == 2 {
		statement = fmt.Sprintf("%s&%s", statement, paginateStatement)
	} else {
		statement = fmt.Sprintf("%s?%s", statement, paginateStatement)
	}

	data, count, err := s.request(statement, withCount)
	if err != nil {
		return data, 0, nil, nil, err
	}

	var nextCursor, prevCursor any
	if cursor != nil {
		prevCursor = s.extractPrevCursor(cursorRefColumn, data)
	}
	nextCursor = s.extractNextCursor(cursorRefColumn, data)

	return data, count, nextCursor, prevCursor, nil
}

func (s *SupabaseDriver) CursorPaginatePrev(ctx context.Context, statement string, cursorRefColumn string, cursor any, limit int, withCount bool) ([]Item, int, any, any, error) {
	paginateStatement := fmt.Sprintf("limit=%d", limit)
	if cursor != nil {
		if reflect.TypeOf(cursor).Kind() == reflect.String && len(cursor.(string)) > 0 {
			paginateStatement = fmt.Sprintf("%s=lt.%v&limit=%d", cursorRefColumn, cursor, limit)
		}
		if reflect.TypeOf(cursor).Kind() == reflect.Int && cursor.(int) > 0 {
			paginateStatement = fmt.Sprintf("%s=lt.%v&limit=%d", cursorRefColumn, cursor, limit)
		}
	}

	if len(strings.Split(statement, "?")) == 2 {
		statement = fmt.Sprintf("%s&%s", statement, paginateStatement)
	} else {
		statement = fmt.Sprintf("%s?%s", statement, paginateStatement)
	}

	data, count, err := s.request(statement, withCount)
	if err != nil {
		return data, 0, nil, nil, err
	}
	nextCursor := s.extractNextCursor(cursorRefColumn, data)
	prevCursor := s.extractPrevCursor(cursorRefColumn, data)
	return data, count, nextCursor, prevCursor, nil
}

func (s *SupabaseDriver) request(statement string, withCount bool) ([]Item, int, error) {
	var count int
	var data []Item

	var url string

	if strings.HasPrefix(s.baseUrl, "/") {
		url = fmt.Sprintf("%s%s", s.baseUrl, statement)
	} else {
		url = fmt.Sprintf("%s/%s", s.baseUrl, statement)
	}

	reqInterceptor := func(req *fasthttp.Request) error {
		if s.token != "" {
			req.Header.Set("Authorization", s.token)
		}

		if s.apiKey != "" {
			req.Header.Set("apikey", s.apiKey)
		}

		if withCount {
			req.Header.Set("Prefer", "count=estimated")
		}

		return nil
	}

	resInterceptor := func(res *fasthttp.Response) error {
		contentRange := string(res.Header.Peek("Content-Range"))
		if contentRange == "" {
			contentRange = string(res.Header.Peek("content-range"))
		}

		if withCount && contentRange != "" {
			parts := strings.Split(contentRange, "/")
			if len(parts) == 2 {
				countInt, err := strconv.Atoi(parts[1])
				if err != nil {
					return err
				}
				count = countInt
			}
		}
		return nil
	}

	body, err := client.SendRequest(fasthttp.MethodGet, url, nil, time.Second*5, reqInterceptor, resInterceptor)
	if err != nil {
		return nil, 0, err
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return nil, 0, err
	}

	return data, count, nil
}

func (s SupabaseDriver) extractNextCursor(cursorRefColumn string, data []Item) any {
	if len(data) == 0 {
		return nil
	}
	lastItem := data[len(data)-1]
	if cursorRef, ok := lastItem[cursorRefColumn].(float64); ok {
		return cursorRef
	}
	return nil
}

func (s SupabaseDriver) extractPrevCursor(cursorRefColumn string, data []Item) any {
	if len(data) == 0 {
		return ""
	}
	firstItem := data[0]
	if cursorRef, ok := firstItem[cursorRefColumn]; ok {
		return cursorRef
	}
	return nil
}
