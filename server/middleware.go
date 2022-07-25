package server

import "sync"

var middlewareContextPool sync.Pool = sync.Pool{
	New: func() any {
		return new(middlewareContext)
	},
}

type middlewareContext struct {
	request  Request
	response Response
	next     func()
	err      ServiceError
}

type MiddlewareContext interface {
	Request() Request
	Response() Response
	Next()
	Report(ServiceError)
	Error() ServiceError
	recycle()
}

func newMiddlewareContext(request Request, response Response) *middlewareContext {
	ctx := middlewareContextPool.Get().(*middlewareContext)
	ctx.request = request
	ctx.response = response
	return ctx
}

func (c *middlewareContext) recycle() {
	middlewareContextPool.Put(c)
}

func (c *middlewareContext) Request() Request {
	return c.request
}

func (c *middlewareContext) Response() Response {
	if c.response == nil {
		c.response = NewResponse(-1, nil)
	}
	return c.response
}

func (c *middlewareContext) Next() {
	c.next()
}

func (c *middlewareContext) Report(err ServiceError) {
	c.err = err
}

func (c *middlewareContext) Error() ServiceError {
	return c.err
}

type Middleware func(ctx MiddlewareContext)

func middlewaresToRequestHandler(middlewares []Middleware) RequestHandler {
	return func(r Request) (Response, ServiceError) {
		return runMiddlewares(middlewares, r)
	}
}

func runMiddlewares(middlewares []Middleware, request Request) (Response, ServiceError) {
	ctx := makeMiddlewareContext(middlewares, request)
	ctx.Next()
	defer func() {
		ctx.recycle()
	}()
	return ctx.Response(), ctx.Error()
}

func makeMiddlewareContext(middlewares []Middleware, request Request) MiddlewareContext {
	response := NewResponse(0, nil)
	currIndex := 0
	ctx := newMiddlewareContext(request, response)
	nextFunc := func() {
		var currMiddleware Middleware
		if ctx.err != nil || currIndex >= len(middlewares) {
			return
		}
		if currIndex < len(middlewares) {
			currMiddleware = middlewares[currIndex]
		}
		currIndex++
		currMiddleware(ctx)
	}
	ctx.next = nextFunc
	return ctx
}

func wrapHandlerAsMiddleware(handler RequestHandler) Middleware {
	return func(ctx MiddlewareContext) {
		rawCtx := ctx.(*middlewareContext)
		rawCtx.response, rawCtx.err = handler(rawCtx.request)
	}
}
