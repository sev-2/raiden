package raiden

// This is a modified version of https://github.com/zeromicro/go-zero/blob/master/rest/chain/chain.go
type (
	// Chain defines a chain of middleware.
	Chain interface {
		Append(middlewares ...MiddlewareFn) Chain
		Prepend(middlewares ...MiddlewareFn) Chain
		Then(fn Controller) RouteHandlerFn
	}

	// chain acts as a list of http.Handler middlewares.
	// chain is effectively immutable:
	// once created, it will always hold
	// the same set of middlewares in the same order.
	chain struct {
		middlewares []MiddlewareFn
	}
)

// New creates a new Chain, memorizing the given list of middleware middlewares.
// New serves no other function, middlewares are only called upon a call to Then() or ThenFunc().
func NewChain(middlewares ...MiddlewareFn) Chain {
	return chain{middlewares: append(([]MiddlewareFn)(nil), middlewares...)}
}

// Append extends a chain, adding the specified middlewares as the last ones in the request flow.
//
//	c := chain.New(m1, m2)
//	c.Append(m3, m4)
//	// requests in c go m1 -> m2 -> m3 -> m4
func (c chain) Append(middlewares ...MiddlewareFn) Chain {
	return chain{middlewares: join(c.middlewares, middlewares)}
}

// Prepend extends a chain by adding the specified chain as the first one in the request flow.
//
//	c := chain.New(m3, m4)
//	c1 := chain.New(m1, m2)
//	c.Prepend(c1)
//	// requests in c go m1 -> m2 -> m3 -> m4
func (c chain) Prepend(middlewares ...MiddlewareFn) Chain {
	return chain{middlewares: join(middlewares, c.middlewares)}
}

// Then chains the middleware and returns the final http.Handler.
//
//	New(m1, m2, m3).Then(h)
//
// is equivalent to:
//
//	m1(m2(m3(h)))
//
// When the request comes in, it will be passed to m1, then m2, then m3
// and finally, the given handler
// (assuming every middleware calls the following one).
func (c chain) Then(controller Controller) RouteHandlerFn {
	handler := BuildHandler(controller)
	for i := range c.middlewares {
		handler = c.middlewares[len(c.middlewares)-1-i](handler)
	}
	return handler
}

func join(a, b []MiddlewareFn) []MiddlewareFn {
	mids := make([]MiddlewareFn, 0, len(a)+len(b))
	mids = append(mids, a...)
	mids = append(mids, b...)
	return mids
}
