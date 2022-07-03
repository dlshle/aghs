package server

import (
	"fmt"
	"net"
	"net/http"

	"github.com/dlshle/gommon/logger"
	"github.com/dlshle/gommon/uri_trie"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

type HTTPServer struct {
	addr                  string
	uriTrie               *uri_trie.TrieTree // trie of <routePattern, RequestHandler>
	services              map[string]Service
	middlewareManager     MiddlewareManager
	logger                logger.Logger
	attachContextForError bool
}

func NewHTTPServer(addr string) HTTPServer {
	return HTTPServer{
		addr:                  addr,
		uriTrie:               uri_trie.NewTrieTree(),
		services:              make(map[string]Service),
		middlewareManager:     NewMiddlewareManager(),
		logger:                logger.GlobalLogger.WithPrefix("[HTTPServer]"),
		attachContextForError: false,
	}
}

func (s HTTPServer) ShouldAttachContextOnServiceError(attach bool) {
	s.attachContextForError = attach
}

func (s HTTPServer) Use(middleware Middleware) {
	s.middlewareManager.RegisterMiddleware(middleware)
}

func (s HTTPServer) UseMiddlewares(middlewares ...Middleware) {
	for _, m := range middlewares {
		s.middlewareManager.RegisterMiddleware(m)
	}
}

func (s HTTPServer) RegisterService(service Service) error {
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

func (s HTTPServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	err := s.HandleHTTP(w, req)
	if err != nil {
		s.logger.Errorf("server encountered an error while handling request(%s, %s) from %s due to %s", req.Method, req.URL.Path, req.RemoteAddr, err.Error())
	}
}

func (s HTTPServer) HandleHTTP(w http.ResponseWriter, req *http.Request) (err error) {
	defer func() {
		// in case of any panic
		if recoveredPanic := recover(); recoveredPanic != nil {
			err = fmt.Errorf("%v", recoveredPanic)
			s.respondWithError(w, InternalError(err.Error()), nil, nil)
		}
	}()
	uri := req.RequestURI
	matchCtx, err := s.uriTrie.Match(uri)
	if err != nil {
		return s.respondWithError(w, NotFoundError(fmt.Sprintf("route %s is undefined", uri)), nil, nil)
	}
	request := s.buildRequest(req, matchCtx)
	traceID := request.Id()
	s.logger.Infof("[%s] receive request %s", traceID, request.String())
	middlewareCtx := s.middlewareManager.Run(request, matchCtx.Value.(Service).Handle)
	resp, serviceErr := middlewareCtx.Response(), middlewareCtx.Error()
	if serviceErr != nil {
		return s.respondWithError(w, serviceErr, resp, request.Context())
	}
	// service handler(core handler) will set middleware.Ctx.response to nil when it doesn't have proper handler operation?
	if resp == nil {
		return s.respondWithError(w, InternalError("invalid handler operation"), resp, request.Context())
	}
	err = s.respondWithServiceResponse(w, resp)
	if err != nil {
		return s.respondWithError(w, InternalError(err.Error()), resp, request.Context())
	}
	return err
}

func (s HTTPServer) buildRequest(r *http.Request, matchCtx *uri_trie.MatchContext) Request {
	return NewRequest(r, matchCtx.Value.(Service), matchCtx.UriPattern, matchCtx.QueryParams, matchCtx.PathParams)
}

func (s HTTPServer) respondWithError(w http.ResponseWriter, serviceErr ServiceError, resp Response, requestCtx RequestContext) (err error) {
	if s.attachContextForError {
		serviceErr.AttachContext(requestCtx)
	}
	w.Header().Set("Content-Type", serviceErr.ContentType())
	w.WriteHeader(serviceErr.Code())
	if resp != nil {
		resp.IterateHeaders(func(k, v string) {
			w.Header().Set(k, v)
		})
	}
	_, err = w.Write([]byte(serviceErr.Error()))
	return
}

func (s HTTPServer) respondWithServiceResponse(w http.ResponseWriter, r Response) (err error) {
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

func (s HTTPServer) Start() error {
	addr := s.addr
	s.logger.Infof("starting the server on %s with TCP protocol...", addr)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		s.logger.Errorf("error starting server at addr %s: %s", addr, err.Error())
		return err
	}
	return fasthttp.Serve(listener, fasthttpadaptor.NewFastHTTPHandler(s))
}
