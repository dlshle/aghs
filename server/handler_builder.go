package server

import "net/http"

type HandlersWithPath interface {
	Path() string
	Handlers() map[string][]Middleware
}

type pathHandlerBuilder struct {
	path     string
	handlers map[string][]Middleware // method:handler
}

func PathHandlerBuilder(path string) *pathHandlerBuilder {
	return &pathHandlerBuilder{
		path:     path,
		handlers: make(map[string][]Middleware),
	}
}

func (b *pathHandlerBuilder) Path() string {
	return b.path
}

func (b *pathHandlerBuilder) Handlers() map[string][]Middleware {
	return b.handlers
}

func (b *pathHandlerBuilder) Get(handler RequestHandler) *pathHandlerBuilder {
	b.handlers[http.MethodGet] = []Middleware{wrapHandlerAsMiddleware(handler)}
	return b
}

func (b *pathHandlerBuilder) GetWithMiddlewares(handler RequestHandler, middlewares ...Middleware) *pathHandlerBuilder {
	b.handlers[http.MethodGet] = append(middlewares, wrapHandlerAsMiddleware(handler))
	return b
}

func (b *pathHandlerBuilder) Post(handler RequestHandler) *pathHandlerBuilder {
	b.handlers[http.MethodPost] = []Middleware{wrapHandlerAsMiddleware(handler)}
	return b
}

func (b *pathHandlerBuilder) PostWithMiddlewares(handler RequestHandler, middlewares ...Middleware) *pathHandlerBuilder {
	b.handlers[http.MethodPost] = append(middlewares, wrapHandlerAsMiddleware(handler))
	return b
}

func (b *pathHandlerBuilder) Patch(handler RequestHandler) *pathHandlerBuilder {
	b.handlers[http.MethodPatch] = []Middleware{wrapHandlerAsMiddleware(handler)}
	return b
}

func (b *pathHandlerBuilder) PatchWithMiddlewares(handler RequestHandler, middlewares ...Middleware) *pathHandlerBuilder {
	b.handlers[http.MethodPatch] = append(middlewares, wrapHandlerAsMiddleware(handler))
	return b
}

func (b *pathHandlerBuilder) Delete(handler RequestHandler) *pathHandlerBuilder {
	b.handlers[http.MethodDelete] = []Middleware{wrapHandlerAsMiddleware(handler)}
	return b
}

func (b *pathHandlerBuilder) DeleteWithMiddlewares(handler RequestHandler, middlewares ...Middleware) *pathHandlerBuilder {
	b.handlers[http.MethodDelete] = append(middlewares, wrapHandlerAsMiddleware(handler))
	return b
}

func (b *pathHandlerBuilder) Put(handler RequestHandler) *pathHandlerBuilder {
	b.handlers[http.MethodPut] = []Middleware{wrapHandlerAsMiddleware(handler)}
	return b
}

func (b *pathHandlerBuilder) PutWithMiddlewares(handler RequestHandler, middlewares ...Middleware) *pathHandlerBuilder {
	b.handlers[http.MethodPut] = append(middlewares, wrapHandlerAsMiddleware(handler))
	return b
}

func (b *pathHandlerBuilder) Head(handler RequestHandler) *pathHandlerBuilder {
	b.handlers[http.MethodHead] = []Middleware{wrapHandlerAsMiddleware(handler)}
	return b
}

func (b *pathHandlerBuilder) HeadWithMiddlewares(handler RequestHandler, middlewares ...Middleware) *pathHandlerBuilder {
	b.handlers[http.MethodHead] = append(middlewares, wrapHandlerAsMiddleware(handler))
	return b
}

func (b *pathHandlerBuilder) Connect(handler RequestHandler) *pathHandlerBuilder {
	b.handlers[http.MethodConnect] = []Middleware{wrapHandlerAsMiddleware(handler)}
	return b
}

func (b *pathHandlerBuilder) ConnectWithMiddlewares(handler RequestHandler, middlewares ...Middleware) *pathHandlerBuilder {
	b.handlers[http.MethodConnect] = append(middlewares, wrapHandlerAsMiddleware(handler))
	return b
}

func (b *pathHandlerBuilder) Options(handler RequestHandler) *pathHandlerBuilder {
	b.handlers[http.MethodOptions] = []Middleware{wrapHandlerAsMiddleware(handler)}
	return b
}

func (b *pathHandlerBuilder) OptionsWithMiddlewares(handler RequestHandler, middlewares ...Middleware) *pathHandlerBuilder {
	b.handlers[http.MethodOptions] = append(middlewares, wrapHandlerAsMiddleware(handler))
	return b
}

func (b *pathHandlerBuilder) Trace(handler RequestHandler) *pathHandlerBuilder {
	b.handlers[http.MethodTrace] = []Middleware{wrapHandlerAsMiddleware(handler)}
	return b
}

func (b *pathHandlerBuilder) TraceWithMiddlewares(handler RequestHandler, middlewares ...Middleware) *pathHandlerBuilder {
	b.handlers[http.MethodTrace] = append(middlewares, wrapHandlerAsMiddleware(handler))
	return b
}

func (b *pathHandlerBuilder) Build() HandlersWithPath {
	return b
}
