package server

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/dlshle/gommon/logger"
)

type Service interface {
	Id() string
	Handle(request Request) (resp Response, err ServiceError)
	UriPatterns() []string
	SupportsRoutePattern(routePattern string) bool
	SupportsMethodForPattern(routePattern, method string) bool
	SupportedMethodsForPattern(pattern string) []string
	Logger() logger.Logger
}

type immutableService struct {
	id          string
	uriMap      map[string]map[string]RequestHandler
	logger      logger.Logger
	middlewares []Middleware
}

func (s immutableService) getRequestHandlingMiddlewares(routePattern, method string) RequestHandler {
	methodMap := s.uriMap[routePattern]
	if methodMap != nil {
		return methodMap[method]
	}
	return nil
}

func (s immutableService) Handle(request Request) (resp Response, err ServiceError) {
	handler := s.getRequestHandlingMiddlewares(request.UriPattern(), request.Method())
	if handler == nil {
		err = MethodNotAllowedError(fmt.Sprintf("method %s is not allowed for uri pattern %s", request.Method(), request.UriPattern()))
		return
	}
	return handler(request)
}

func (s immutableService) UriPatterns() []string {
	var patterns []string
	for pattern, _ := range s.uriMap {
		patterns = append(patterns, pattern)
	}
	return patterns
}

func (s immutableService) Id() string {
	return s.id
}

func (s immutableService) SupportsRoutePattern(routePattern string) bool {
	return s.uriMap[routePattern] != nil
}

func (s immutableService) SupportsMethodForPattern(routePattern, method string) bool {
	return s.SupportsRoutePattern(routePattern) && s.uriMap[routePattern][method] != nil
}

func (s immutableService) SupportedMethodsForPattern(pattern string) []string {
	var supportedMethods []string
	if s.uriMap[pattern] == nil {
		return supportedMethods
	}
	for method := range s.uriMap[pattern] {
		supportedMethods = append(supportedMethods, method)
	}
	return supportedMethods
}

func (s immutableService) Logger() logger.Logger {
	return s.logger
}

type ServiceBuilder interface {
	Id(string) ServiceBuilder
	Middlewares(...Middleware) ServiceBuilder
	WithRouteHandlers(path HandlersWithPath) ServiceBuilder
	LogWriter(io.Writer) ServiceBuilder
	Build() (Service, error)
}

type immutableServiceBuilder struct {
	s           *immutableService
	uriMap      map[string]map[string][]Middleware
	middlewares []Middleware
	err         error
	writer      io.Writer
}

func NewServiceBuilder() ServiceBuilder {
	return &immutableServiceBuilder{
		s: &immutableService{
			uriMap:      make(map[string]map[string]RequestHandler),
			middlewares: make([]Middleware, 0),
		},
		uriMap:      make(map[string]map[string][]Middleware),
		middlewares: make([]Middleware, 0),
		err:         nil,
		writer:      nil,
	}
}

func (b *immutableServiceBuilder) Middlewares(middlewares ...Middleware) ServiceBuilder {
	if b.err != nil {
		return b
	}
	b.middlewares = middlewares
	return b
}

func (b *immutableServiceBuilder) WithRouteHandlers(handlersWithPath HandlersWithPath) ServiceBuilder {
	if b.err != nil {
		return b
	}
	path := handlersWithPath.Path()
	handlers := handlersWithPath.Handlers()
	if b.uriMap[path] != nil {
		b.err = fmt.Errorf("path %s has already been registered", path)
		return b
	}
	b.uriMap[path] = handlers
	return b
}

func (b *immutableServiceBuilder) Build() (Service, error) {
	if b.err != nil {
		return nil, b.err
	}
	if b.s.id == "" {
		return nil, fmt.Errorf("service id is empty")
	}
	if len(b.uriMap) == 0 {
		return nil, fmt.Errorf("service has empty handler map")
	}
	b.tryConcatServiceMiddlewaresToRequestHandlerMiddlewares()
	b.compressRequestHandlerMiddlewares()
	return b.s, nil
}

func (b *immutableServiceBuilder) tryConcatServiceMiddlewaresToRequestHandlerMiddlewares() {
	if b.middlewares == nil {
		return
	}
	for u, v := range b.uriMap {
		for k, m := range v {
			b.uriMap[u][k] = append(b.middlewares, m...)
		}
	}
}

func (b *immutableServiceBuilder) compressRequestHandlerMiddlewares() {
	for u, v := range b.uriMap {
		requestHandlerMap := make(map[string]RequestHandler)
		for k, m := range v {
			requestHandlerMap[k] = middlewaresToRequestHandler(m)
		}
		b.s.uriMap[u] = requestHandlerMap
	}
}

func (b *immutableServiceBuilder) Id(id string) ServiceBuilder {
	b.s.id = id
	var writer io.Writer
	if b.writer != nil {
		writer = b.writer
	} else {
		writer = os.Stdout
	}
	b.s.logger = logger.NewLevelLogger(writer, fmt.Sprintf("[service-%s]", id), log.Ldate|log.Ltime, logger.TRACE)
	return b
}

func (b *immutableServiceBuilder) LogWriter(writer io.Writer) ServiceBuilder {
	if b.s.id != "" {
		b.s.logger = logger.GlobalLogger.WithPrefix("[service-" + b.s.id + "]")
	}
	b.writer = writer
	return b
}
