package paginate

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sev-2/raiden"
)

// PaginationType defines the type of pagination
type Type string

const (
	OffsetPagination Type = "offset"
	CursorPagination Type = "cursor"
)

type CursorPaginateDirection string

const (
	CursorPaginateDirectionNext CursorPaginateDirection = "next"
	CursorPaginateDirectionPrev CursorPaginateDirection = "prev"
)

const DefaultOffsetColumn = "id"

type Executor interface {
	Execute(ctx context.Context, statement string) (ExecuteResult[Item], error)
	SetDriver(driver Driver) Executor
}

type ExecuteOptions struct {
	Page  int
	Limit int
	Type  Type

	Cursor          any
	CursorDirection CursorPaginateDirection
	CursorRefColumn string

	WithCount bool
	IsBypass  bool
}

type ExecuteResult[T any] struct {
	Data       []T
	Count      int
	NextCursor any
	PrevCursor any
}

type executor struct {
	options ExecuteOptions
	driver  Driver
}

func New(config *raiden.Config, opts ExecuteOptions) Executor {
	driver := &SupabaseDriver{}

	baseUrl := config.SupabasePublicUrl
	if config.Mode == raiden.SvcMode {
		baseUrl = config.PostgRestUrl
	}
	baseUrl = strings.TrimSuffix(baseUrl, "/")

	if config.Mode == raiden.BffMode {
		baseUrl = fmt.Sprintf("%s/rest/v1", baseUrl)
	}

	driver.baseUrl = baseUrl
	driver.apiKey = config.ServiceKey
	driver.token = fmt.Sprintf("Bearer %s", config.ServiceKey)

	if config.Mode == raiden.SvcMode && config.JwtToken != "" {
		driver.token = fmt.Sprintf("Bearer %s", config.JwtToken)
	}

	return &executor{
		options: opts,
		driver:  driver,
	}
}

func NewFromContext(ctx raiden.Context, opts ExecuteOptions) Executor {
	driver := &SupabaseDriver{}

	baseUrl := ctx.Config().SupabasePublicUrl
	if ctx.Config().Mode == raiden.SvcMode {
		baseUrl = ctx.Config().PostgRestUrl
	}
	baseUrl = strings.TrimSuffix(baseUrl, "/")

	if ctx.Config().Mode == raiden.BffMode {
		baseUrl = fmt.Sprintf("%s/rest/v1", baseUrl)
	}
	driver.baseUrl = baseUrl

	if opts.IsBypass {
		driver.apiKey = ctx.Config().ServiceKey
		driver.token = fmt.Sprintf("Bearer %s", ctx.Config().ServiceKey)

		if ctx.Config().Mode == raiden.SvcMode && ctx.Config().JwtToken != "" {
			driver.token = fmt.Sprintf("Bearer %s", ctx.Config().JwtToken)
		}
	} else {
		apikey := string(ctx.RequestContext().Request.Header.Peek("apikey"))
		if apikey != "" {
			driver.apiKey = apikey
		}
		bearerToken := string(ctx.RequestContext().Request.Header.Peek("Authorization"))
		if bearerToken != "" && strings.HasPrefix(bearerToken, "Bearer ") {
			driver.token = bearerToken
		}
	}

	return &executor{
		options: opts,
		driver:  driver,
	}
}

func (e *executor) Execute(ctx context.Context, statement string) (ExecuteResult[Item], error) {
	var result ExecuteResult[Item]

	if e.options.Type == OffsetPagination {
		data, count, err := e.driver.Paginate(ctx, statement, e.options.Page, e.options.Limit, e.options.WithCount)
		if err != nil {
			return result, err
		}

		result.Data = data
		result.Count = count
	}

	if e.options.Type == CursorPagination {
		cursorRefColumn := e.options.CursorRefColumn
		if cursorRefColumn == "" {
			cursorRefColumn = "id"
		}

		if e.options.CursorDirection == CursorPaginateDirectionNext {
			data, count, nextCursor, prevCursor, err := e.driver.CursorPaginateNext(ctx, statement, cursorRefColumn, e.options.Cursor, e.options.Limit, e.options.WithCount)
			if err != nil {
				return result, err
			}

			result.Data = data
			result.Count = count
			result.NextCursor = nextCursor
			result.PrevCursor = prevCursor
		}

		if e.options.CursorDirection == CursorPaginateDirectionPrev {
			data, count, nextCursor, prevCursor, err := e.driver.CursorPaginatePrev(ctx, statement, cursorRefColumn, e.options.Cursor, e.options.Limit, e.options.WithCount)
			if err != nil {
				return result, err
			}

			result.Data = data
			result.Count = count
			result.NextCursor = nextCursor
			result.PrevCursor = prevCursor
		}
	}

	if result.Data == nil {
		result.Data = make([]Item, 0)
	}
	return result, nil
}

func (e *executor) SetDriver(driver Driver) Executor {
	e.driver = driver
	return e
}

func MarshallResult[T any](data ExecuteResult[Item]) (result ExecuteResult[T], err error) {
	byteData, err := json.Marshal(data)
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(byteData, &result)
	return
}

func SendResponse[T any](ctx raiden.Context, data ExecuteResult[T]) error {
	ctx.RequestContext().Response.Header.Add("Content-Range", fmt.Sprintf("0-24/%d", data.Count))
	if data.NextCursor != nil {
		ctx.RequestContext().Response.Header.Add("Next-Cursor", fmt.Sprintf("%v", data.NextCursor))
	}

	if data.PrevCursor != nil {
		ctx.RequestContext().Response.Header.Add("Prev-Cursor", fmt.Sprintf("%v", data.PrevCursor))
	}
	return ctx.SendJson(data.Data)
}
