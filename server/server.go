package server

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/dlshle/aghs/logger"
	"github.com/dlshle/aghs/uri_trie"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

type Server struct {
	uriTrie               *uri_trie.TrieTree // trie of <routePattern, RequestHandler>
	services              map[string]Service
	middlewareManager     MiddlewareManager
	logger                logger.Logger
	attachContextForError bool
}

func New() Server {
	return Server{
		uriTrie:               uri_trie.NewTrieTree(),
		services:              make(map[string]Service),
		middlewareManager:     NewMiddlewareManager(),
		logger:                logger.NewLevelLogger(os.Stdout, "[Server]", log.Ldate|log.Ltime, logger.TRACE),
		attachContextForError: false,
	}
}

func (s Server) ShouldAttachContextOnServiceError(attach bool) {
	s.attachContextForError = attach
}

func (s Server) Use(middleware Middleware) {
	s.middlewareManager.RegisterMiddleware(middleware)
}

func (s Server) UseMiddlewares(middlewares ...Middleware) {
	for _, m := range middlewares {
		s.middlewareManager.RegisterMiddleware(m)
	}
}

func (s Server) RegisterService(service Service) error {
	if _, exists := s.services[service.Id()]; exists {
		return fmt.Errorf("service %s already exists", service.Id())
	}
	s.services[service.Id()] = service
	for _, pattern := range service.UriPatterns() {
		err := s.uriTrie.Add(pattern, service, true)
		if err != nil {
			s.logger.Errorf("error while adding route %s from service %s: %s", pattern, service.Id(), err.Error())
			return err
		}
	}
	s.logger.Infof("service %s has been registered", service.Id())
	return nil
}

func (s Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	err := s.HandleHTTP(w, req)
	if err != nil {
		s.logger.Errorf("server encountered an error while handling request(%s, %s) from %s due to %s", req.Method, req.URL.Path, req.RemoteAddr, err.Error())
	}
}

func (s Server) HandleHTTP(w http.ResponseWriter, req *http.Request) (err error) {
	defer func() {
		// in case of any panic
		if recoveredPanic := recover(); recoveredPanic != nil {
			err = fmt.Errorf("%v", recoveredPanic)
			s.respondWithError(w, InternalError(err.Error()), nil)
		}
	}()
	uri := req.RequestURI
	matchCtx, err := s.uriTrie.Match(uri)
	if err != nil {
		return s.respondWithError(w, NotFoundError(fmt.Sprintf("route %s is undefined", uri)), nil)
	}
	request := s.buildRequest(req, matchCtx)
	traceId := request.Id()
	s.logger.Infof("[%s] receive request %s", traceId, request.String())
	middlewareCtx := s.middlewareManager.Run(request, matchCtx.Value.(Service).Handle)
	resp, serviceErr := middlewareCtx.Response(), middlewareCtx.Error()
	if serviceErr != nil {
		return s.respondWithError(w, serviceErr, request.Context())
	}
	// service handler(core handler) will set middleware.Ctx.response to nil when it doesn't have proper handler operation?
	if resp == nil {
		return s.respondWithError(w, InternalError("invalid handler operation"), request.Context())
	}
	err = s.respondWithServiceResponse(w, resp)
	if err != nil {
		return s.respondWithError(w, InternalError(err.Error()), request.Context())
	}
	return err
}

func (s Server) buildRequest(r *http.Request, matchCtx *uri_trie.MatchContext) Request {
	return NewRequest(r, matchCtx.UriPattern, matchCtx.QueryParams, matchCtx.PathParams)
}

func (s Server) respondWithError(w http.ResponseWriter, serviceErr ServiceError, requestCtx RequestContext) (err error) {
	if s.attachContextForError {
		serviceErr.AttachContext(requestCtx)
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(serviceErr.Code())
	_, err = w.Write([]byte(serviceErr.Error()))
	return
}

func (s Server) respondWithServiceResponse(w http.ResponseWriter, r *Response) (err error) {
	if r.Code() == 0 {
		return fmt.Errorf("invalid payload")
	}
	w.Header().Set("Content-Type", r.ContentType())
	w.WriteHeader(r.Code())
	r.IterateHeaders(func(k string, v string) {
		w.Header().Set(k, v)
	})
	stream, err := r.PayloadStream()
	if err != nil {
		return err
	}
	_, err = w.Write(stream)
	return
}

func (s Server) Start(addr string) error {
	s.logger.Infof("starting the server on %s with TCP protocol...", addr)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		s.logger.Errorf("error starting server at addr %s: %s", addr, err.Error())
		return err
	}
	err = fasthttp.Serve(listener, fasthttpadaptor.NewFastHTTPHandler(s))
	if err != nil {
		return err
	}
	s.logger.Info("server has started on ", addr)
	return nil
}
