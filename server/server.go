package server

import (
	"fmt"
	"net"
	"net/http"

	"github.com/dlshle/gommon/logger"
	"github.com/dlshle/gommon/uri_trie"
)

type Server interface {
	Start() error
}

type immutableServer struct {
	engine                Engine
	addr                  string
	uriTrie               *uri_trie.TrieTree
	middlewares           []Middleware
	logger                logger.Logger
	attachContextForError bool
}

func (s immutableServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	err := s.HandleHTTP(w, req)
	if err != nil {
		s.logger.Errorf("server encountered an error while handling request(%s, %s) from %s due to %s", req.Method, req.URL.Path, req.RemoteAddr, err.Error())
	}
}

func (s immutableServer) HandleHTTP(w http.ResponseWriter, req *http.Request) (err error) {
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
	serverRequest := s.buildRequest(req, matchCtx)
	traceID := serverRequest.Id()
	s.logger.Infof("[%s] receive request %s", traceID, serverRequest.String())
	middlewares := append(s.middlewares, wrapHandlerAsMiddleware(matchCtx.Value.(Service).Handle))
	resp, serviceErr := runMiddlewares(middlewares, serverRequest)
	defer func() {
		// matchCtx.Recycle()
		serverRequest.(*request).recycle()
		resp.(*response).recycle()
	}()
	if serviceErr != nil {
		return s.respondWithError(w, serviceErr, resp, serverRequest.Context())
	}
	// service handler(core handler) will set middleware.Ctx.response to nil when it doesn't have proper handler operation?
	if resp == nil {
		return s.respondWithError(w, InternalError("invalid handler operation"), resp, serverRequest.Context())
	}
	err = s.respondWithServiceResponse(w, resp)
	if err != nil {
		return s.respondWithError(w, InternalError(err.Error()), resp, serverRequest.Context())
	}
	return err
}

func (s immutableServer) buildRequest(r *http.Request, matchCtx *uri_trie.MatchContext) Request {
	return NewRequest(r, matchCtx.Value.(Service), matchCtx.UriPattern, matchCtx.QueryParams, matchCtx.PathParams)
}

func (s immutableServer) respondWithError(w http.ResponseWriter, serviceErr ServiceError, resp Response, requestCtx RequestContext) (err error) {
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

func (s immutableServer) respondWithServiceResponse(w http.ResponseWriter, r Response) (err error) {
	if r.Code() == 0 {
		return fmt.Errorf("invalid payload")
	}
	if r.ContentType() != "" {
		w.Header().Set("Content-Type", r.ContentType())
	}
	r.IterateHeaders(func(k string, v string) {
		w.Header().Set(k, v)
	})
	w.WriteHeader(r.Code())
	if r.Code() == http.StatusNoContent {
		return
	}
	stream, err := r.PayloadStream()
	if err != nil {
		return err
	}
	_, err = w.Write(stream)
	return
}

func (s immutableServer) Start() error {
	addr := s.addr
	s.logger.Infof("starting the server on %s with TCP protocol...", addr)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		s.logger.Errorf("error starting server at addr %s: %s", addr, err.Error())
		return err
	}
	return s.startEngine(listener)
}

func (s immutableServer) startEngine(listener net.Listener) error {
	return s.engine(listener, s)
}

type Builder interface {
	Engine(engine Engine) Builder
	Address(string) Builder
	WithServices([]Service) Builder
	WithService(Service) Builder
	WithMiddlewares([]Middleware) Builder
	WithMiddleware(Middleware) Builder
	Logger(logger.Logger) Builder
	AttachContextForError(bool) Builder
	Build() (Server, error)
}

type serverBuilder struct {
	engine                Engine
	addr                  string
	uriTrie               *uri_trie.TrieTree
	middlewares           []Middleware
	logger                logger.Logger
	attachContextForError bool
	serviceIdSet          map[string]bool
	err                   error
}

func (s *serverBuilder) Engine(engine Engine) Builder {
	s.engine = engine
	return s
}

func (s *serverBuilder) Address(addr string) Builder {
	s.addr = addr
	return s
}

func (s *serverBuilder) WithServices(services []Service) Builder {
	for _, svc := range services {
		if !s.addService(svc) {
			return s
		}
	}
	return s
}

func (s *serverBuilder) WithService(svc Service) Builder {
	s.addService(svc)
	return s
}

func (s *serverBuilder) WithMiddlewares(middlewares []Middleware) Builder {
	s.middlewares = append(s.middlewares, middlewares...)
	return s
}

func (s *serverBuilder) WithMiddleware(middleware Middleware) Builder {
	s.middlewares = append(s.middlewares, middleware)
	return s
}

func (s *serverBuilder) Logger(l logger.Logger) Builder {
	s.logger = l
	return s
}

func (s *serverBuilder) AttachContextForError(attach bool) Builder {
	s.attachContextForError = attach
	return s
}

func (s *serverBuilder) Build() (Server, error) {
	if s.err != nil {
		return nil, s.err
	}
	return immutableServer{
		engine:                s.engine,
		addr:                  s.addr,
		uriTrie:               s.uriTrie,
		middlewares:           s.middlewares,
		logger:                s.logger,
		attachContextForError: s.attachContextForError,
	}, nil
}

func (s *serverBuilder) addService(service Service) bool {
	if _, exists := s.serviceIdSet[service.Id()]; exists {
		s.err = fmt.Errorf("service %s already exists", service.Id())
		return false
	}
	s.serviceIdSet[service.Id()] = true
	for _, pattern := range service.UriPatterns() {
		err := s.uriTrie.Add(pattern, service, true)
		if err != nil {
			s.logger.Errorf("error while adding route %s from service %s: %s", pattern, service.Id(), err.Error())
			s.err = err
			return false
		}
	}
	return true
}

func NewBuilder() Builder {
	return &serverBuilder{
		middlewares:  make([]Middleware, 0),
		serviceIdSet: make(map[string]bool),
		uriTrie:      uri_trie.NewTrieTree(),
		logger:       logger.GlobalLogger.WithPrefix("[HTTPServer]"),
		engine:       NetEngine,
	}
}
