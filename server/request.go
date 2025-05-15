package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/dlshle/aghs/utils"
)

var (
	requestPool sync.Pool = sync.Pool{New: func() any {
		return new(request)
	}}
	requestContextPool sync.Pool = sync.Pool{New: func() any {
		return make(RequestContext)
	}}
)

const (
	ContextKeyUriPattern     = "uri_pattern"
	ContextKeyQueryParams    = "query_params"
	ContextKeyPathParams     = "path_params"
	ContextKeyMatchedService = "matched_service"
)

type Request interface {
	Id() string
	String() string
	UriPattern() string
	Path() string
	URI() string
	PathParams() map[string]string
	QueryParams() map[string]string
	MatchedService() Service
	Method() string
	Header() http.Header
	Body() ([]byte, error)
	FormFile(key string, maxSize int64) (io.ReadCloser, error)
	MultipartFormValue(key string, maxSize int64) ([]string, error)
	FormValue(key string) (string, error)
	UnmarshalBody(holder interface{}) error
	RemoteAddress() string
	GetContext(key string) interface{}
	RegisterContext(key string, value interface{})
	UnRegisterContext(key string) bool
	Context() RequestContext
	RawCtx() context.Context
}

type request struct {
	id   string
	r    *http.Request
	c    RequestContext
	body []byte
}

func NewRequest(r *http.Request, matchedSvc Service, uriPattern string, queryParams map[string]string, pathParams map[string]string) Request {
	c := requestContextPool.Get().(RequestContext)
	c.RegisterContext(ContextKeyUriPattern, uriPattern)
	c.RegisterContext(ContextKeyQueryParams, queryParams)
	c.RegisterContext(ContextKeyPathParams, pathParams)
	c.RegisterContext(ContextKeyMatchedService, matchedSvc)
	request := requestPool.Get().(*request)
	request.id = utils.GenerateID()
	request.r = r
	request.c = c
	return request
}

func (r *request) recycle() {
	r.body = nil
	r.id = ""
	r.r = nil
	requestContextPool.Put(r.c)
	requestPool.Put(r)
}

func (r *request) Id() string {
	return r.id
}

func (r *request) String() string {
	return fmt.Sprintf(`Request{"id":"%s","method":"%s","url":"%s","remoteAddr":"%s","header":"%s","context":"%s","body":"%s"}`,
		r.id,
		r.Method(),
		r.r.URL.String(),
		r.RemoteAddress(),
		r.Header(),
		r.Context(),
		truncateString(r.tryGetBody(), 64),
	)
}

func (r *request) tryGetBody() string {
	body, err := r.Body()
	if err != nil {
		return "ERROR: unable to read body due to " + err.Error()
	}
	return string(body)
}

func (r *request) UriPattern() string {
	return r.GetContext(ContextKeyUriPattern).(string)
}

func (r *request) Path() string {
	return r.r.URL.Path
}

// URI contains query params
func (r *request) URI() string {
	return r.r.RequestURI
}

func (r *request) PathParams() map[string]string {
	return r.GetContext(ContextKeyPathParams).(map[string]string)
}

func (r *request) QueryParams() map[string]string {
	return r.GetContext(ContextKeyQueryParams).(map[string]string)
}

func (r *request) MatchedService() Service {
	return r.GetContext(ContextKeyMatchedService).(Service)
}

func (r *request) Method() string {
	return r.r.Method
}

func (r *request) Header() http.Header {
	return r.r.Header
}

func (r *request) Body() ([]byte, error) {
	if r.body == nil {
		bodyBytes, err := io.ReadAll(r.r.Body)
		if err != nil {
			return nil, err
		}
		r.body = bodyBytes
	}
	return r.body, nil
}

func (r *request) FormFile(key string, maxSize int64) (io.ReadCloser, error) {
	if err := r.r.ParseMultipartForm(10 << 20); err != nil {
		return nil, fmt.Errorf("failed to parse multipart form: %v", err)
	}
	file, _, err := r.r.FormFile(key)
	if err != nil {
		return nil, fmt.Errorf("failed to extract multipart form file: %v", err)
	}
	return file, nil
}

func (r *request) FormValue(key string) (string, error) {
	err := r.r.ParseForm()
	if err != nil {
		return "", err
	}
	return r.r.FormValue(key), nil
}

func (r *request) MultipartFormValue(key string, maxSize int64) ([]string, error) {
	err := r.r.ParseMultipartForm(maxSize)
	if err != nil {
		return nil, err
	}
	return r.r.MultipartForm.Value[key], nil
}

func (r *request) UnmarshalBody(holder interface{}) error {
	bodyStream, err := r.Body()
	if err != nil {
		return err
	}
	return json.Unmarshal(bodyStream, holder)
}

func (r *request) RemoteAddress() string {
	return r.r.RemoteAddr
}

func (r *request) GetContext(key string) interface{} {
	return r.c.GetContext(key)
}

func (r *request) RegisterContext(key string, value interface{}) {
	r.c.RegisterContext(key, value)
}

func (r *request) UnRegisterContext(key string) bool {
	return r.c.UnRegisterContext(key)
}

func (r *request) Context() RequestContext {
	return r.c
}

func (r *request) RawCtx() context.Context {
	return r.r.Context()
}

func truncateString(s string, l int) string {
	if len(s) <= l {
		return s
	}
	return s[0:l]
}
