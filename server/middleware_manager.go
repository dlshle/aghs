package server

import (
	"sync"
)

type MiddlewareManager interface {
	RegisterMiddleware(Middleware)
	Run(Request, RequestHandler) MiddlewareContext
}

type middlewareManager struct {
	middlewares []Middleware
	lock        *sync.RWMutex
}

func NewMiddlewareManager() MiddlewareManager {
	return &middlewareManager{
		middlewares: make([]Middleware, 0),
		lock:        new(sync.RWMutex),
	}
}

func (m *middlewareManager) withWrite(cb func()) {
	m.lock.Lock()
	defer m.lock.Unlock()
	cb()
}

func (m *middlewareManager) withRead(cb func()) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	cb()
}

func (m *middlewareManager) RegisterMiddleware(handler Middleware) {
	m.withWrite(func() {
		m.middlewares = append(m.middlewares, handler)
	})
}

func (m *middlewareManager) Run(request Request, coreHandler RequestHandler) MiddlewareContext {
	response := NewResponse(0, nil)
	ctx := m.makeMiddlewareContext(request, response, coreHandler)
	ctx.Next()
	return ctx
}

func (m *middlewareManager) makeMiddlewareContext(request Request, response *Response, coreHandler RequestHandler) *middlewareContext {
	coreHandlerWrapper := func(ctx MiddlewareContext) {
		rawCtx := ctx.(*middlewareContext)
		rawCtx.response, rawCtx.err = coreHandler(rawCtx.request)
	}
	ctx := &middlewareContext{
		request:  request,
		response: response,
	}
	currIndex := 0
	nextFunc := func() {
		var handler Middleware
		if ctx.err != nil {
			return
		}
		m.withRead(func() {
			if currIndex == len(m.middlewares) {
				handler = coreHandlerWrapper
				return
			}
			if currIndex < len(m.middlewares) {
				handler = m.middlewares[currIndex]
			}
		})
		currIndex++
		handler(ctx)
	}
	ctx.next = nextFunc
	return ctx
}
