package server

import (
	"fmt"
	"github.com/dlshle/aghs/logger"
	"log"
	"os"
	"sync"
)

type MutableService interface {
	Service
	RegisterHandlersByPath(routePattern string, handlers map[string]RequestHandler) (err error)
	RegisterHandler(method, routePattern string, handler RequestHandler) (err error)
	UnRegisterHandler(method, routePattern string) (err error)
	Use(middleware Middleware)
}

type service struct {
	id          string
	uriMap      map[string]map[string][]Middleware
	lock        *sync.RWMutex
	logger      logger.Logger
	middlewares []Middleware
}

func NewService(id string) MutableService {
	return service{
		id:          id,
		uriMap:      make(map[string]map[string][]Middleware),
		lock:        new(sync.RWMutex),
		middlewares: make([]Middleware, 0),
		logger:      logger.NewLevelLogger(os.Stdout, fmt.Sprintf("[service-%s]", id), log.Ldate|log.Ltime, logger.TRACE),
	}
}

func (s service) withWrite(cb func()) {
	s.lock.Lock()
	defer s.lock.Unlock()
	cb()
}

func (s service) Id() string {
	return s.id
}

func (s service) RegisterHandlersByPath(routePattern string, handlers map[string]RequestHandler) (err error) {
	for k, v := range handlers {
		if err = s.RegisterHandler(k, routePattern, v); err != nil {
			return
		}
	}
	return
}

// this overrides the previous handler at [route][method]
func (s service) RegisterHandler(method, routePattern string, handler RequestHandler) (err error) {
	if handler == nil {
		return fmt.Errorf("invalid handler")
	}
	s.withWrite(func() {
		if s.uriMap[routePattern] == nil {
			s.uriMap[routePattern] = make(map[string][]Middleware)
		}
		if s.uriMap[routePattern][method] == nil {
			s.uriMap[routePattern][method] = []Middleware{wrapHandlerAsMiddleware(handler)}
		}
	})
	s.logger.Infof("handler (%s, %s) has been registered", method, routePattern)
	return
}

// don't think this is necessary though
func (s service) UnRegisterHandler(method, routePattern string) (err error) {
	s.withWrite(func() {
		if s.uriMap[routePattern] == nil {
			err = s.notFoundError("routePattern", routePattern)
			return
		}
		if s.uriMap[routePattern][method] == nil {
			err = s.methodNotFoundError(method, routePattern)
			return
		}
		delete(s.uriMap[routePattern], method)
		if len(s.uriMap[routePattern]) == 0 {
			delete(s.uriMap, routePattern)
		}
	})
	return
}

func (s service) UriPatterns() []string {
	var patterns []string
	for pattern, _ := range s.uriMap {
		patterns = append(patterns, pattern)
	}
	return patterns
}

func (s service) SupportsRoutePattern(routePattern string) bool {
	return s.uriMap[routePattern] != nil
}

func (s service) SupportsMethodForPattern(routePattern, method string) bool {
	return s.SupportsRoutePattern(routePattern) && s.uriMap[routePattern][method] != nil
}

func (s service) getRequestHandlingMiddlewares(routePattern, method string) []Middleware {
	methodMap := s.uriMap[routePattern]
	if methodMap != nil {
		return methodMap[method]
	}
	return nil
}

func (s service) Handle(request Request) (resp *Response, err ServiceError) {
	middlewares := s.getRequestHandlingMiddlewares(request.UriPattern(), request.Method())
	if middlewares == nil {
		err = MethodNotAllowedError(fmt.Sprintf("method %s is not allowed for uri pattern %s", request.Method(), request.UriPattern()))
		return
	}
	return runMiddlewares(middlewares, request)
}

func (s service) Use(middleware Middleware) {
	s.withWrite(func() {
		s.middlewares = append(s.middlewares, middleware)
	})
}

func (s service) Logger() logger.Logger {
	return s.logger
}

func (s service) methodNotFoundError(method, routePattern string) error {
	return fmt.Errorf("method %s for routePattern %s does not exist", method, routePattern)
}

func (s service) notFoundError(itemType, item string) error {
	return fmt.Errorf("%s %s does not exist", itemType, item)
}
