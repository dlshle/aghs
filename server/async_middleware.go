package server

import (
	"sync"

	"github.com/dlshle/gommon/async"
)

var asyncMiddlewareCtxPool sync.Pool = sync.Pool{
	New: func() any {
		return &asyncMiddlewareCtx{}
	},
}

type asyncMiddlewareCtx struct {
	request  Request
	response Response
	next     func() async.Future
	err      ServiceError
}

type AsyncMiddleware func(ctx *asyncMiddlewareCtx) async.Future

func newAsyncMiddlewareContext(request Request, response Response) *asyncMiddlewareCtx {
	ctx := middlewareContextPool.Get().(*asyncMiddlewareCtx)
	ctx.request = request
	ctx.response = response
	return ctx
}

func (c *asyncMiddlewareCtx) recycle() {
	asyncMiddlewareCtxPool.Put(c)
}

func (c *asyncMiddlewareCtx) Request() Request {
	return c.request
}

func (c *asyncMiddlewareCtx) Response() Response {
	if c.response == nil {
		c.response = NewResponse(-1, nil)
	}
	return c.response
}

func (c *asyncMiddlewareCtx) Next() async.Future {
	return c.next()
}

func (c *asyncMiddlewareCtx) Report(err ServiceError) {
	c.err = err
}

func (c *asyncMiddlewareCtx) Error() ServiceError {
	return c.err
}

/* fix later
func runAsyncMiddlewares(middlewares []AsyncMiddleware, request Request) async.Future {
	ctx := makeAsyncMiddlewareContext(middlewares, request)
	return ctx.Next().Then(func(i interface{}) interface{} {
		ctx.recycle()
		return nil
	}).OnError(func(err error) {
		ctx.recycle()
	}).OnPanic(func(i interface{}) {
		ctx.recycle()
	})
}
*/

func makeAsyncMiddlewareContext(middlewares []AsyncMiddleware, request Request) *asyncMiddlewareCtx {
	response := NewResponse(0, nil)
	currIndex := 0
	ctx := newAsyncMiddlewareContext(request, response)
	nextFunc := func() async.Future {
		var currMiddleware AsyncMiddleware
		if ctx.err != nil || currIndex >= len(middlewares) {
			return async.ImmediateErrorFuture(ctx.err)
		}
		if currIndex < len(middlewares) {
			currMiddleware = middlewares[currIndex]
		}
		currIndex++
		return currMiddleware(ctx)
	}
	ctx.next = nextFunc
	return ctx
}
