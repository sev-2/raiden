package paginate

import (
	"context"
)

type Item map[string]any

type Driver interface {
	Paginate(ctx context.Context, statement string, page, limit int, withCount bool) (data []Item, cont int, err error)
	CursorPaginateNext(ctx context.Context, statement string, cursorRefColumn string, cursor any, limit int, withCount bool) (data []Item, count int, prevCursor any, nextCursor any, err error)
	CursorPaginatePrev(ctx context.Context, statement string, cursorRefColumn string, cursor any, limit int, withCount bool) (data []Item, count int, prevCursor any, nextCursor any, err error)
}
