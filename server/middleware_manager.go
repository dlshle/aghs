package server

import (
	"sync"
)

type MiddlewareManager struct {
	middlewares []Middleware
	lock        *sync.RWMutex
}

func NewMiddlewareManager() MiddlewareManager {
	return MiddlewareManager{
		middlewares: make([]Middleware, 0),
		lock:        new(sync.RWMutex),
	}
}

func (m *MiddlewareManager) withWrite(cb func()) {
	m.lock.Lock()
	defer m.lock.Unlock()
	cb()
}

func (m *MiddlewareManager) withRead(cb func()) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	cb()
}

func (m *MiddlewareManager) RegisterMiddleware(handler Middleware) {
	m.withWrite(func() {
		m.middlewares = append(m.middlewares, handler)
	})
}

func (m *MiddlewareManager) Run(request Request, coreHandler RequestHandler) MiddlewareContext {
	response := NewResponse(0, nil)
	ctx := m.makeMiddlewareContext(request, response, coreHandler)
	ctx.Next()
	return ctx
}

func (m *MiddlewareManager) makeMiddlewareContext(request Request, response *Response, coreHandler RequestHandler) *middlewareContext {
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
