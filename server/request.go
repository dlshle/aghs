package server

import (
	"encoding/json"
	"fmt"
	"github.com/dlshle/aghs/utils"
	"io/ioutil"
	"net/http"
)

const (
	ContextKeyUriPattern  = "uri_pattern"
	ContextKeyQueryParams = "query_params"
	ContextKeyPathParams  = "path_params"
)

type Request struct {
	id string
	r  *http.Request
	c  RequestContext
}

func NewRequest(r *http.Request, uriPattern string, queryParams map[string]string, pathParams map[string]string) Request {
	c := make(RequestContext)
	c.RegisterContext(ContextKeyUriPattern, uriPattern)
	c.RegisterContext(ContextKeyQueryParams, queryParams)
	c.RegisterContext(ContextKeyPathParams, pathParams)
	return Request{
		id: utils.GenerateID(),
		r:  r,
		c:  c,
	}
}

func (r Request) Id() string {
	return r.id
}

func (r Request) String() string {
	return fmt.Sprintf(`Request{"id":"%s","method":"%s","url":"%s","remoteAddr":"%s","header":"%s","context":"%s"`,
		r.id,
		r.Method(),
		r.r.URL.String(),
		r.RemoteAddress(),
		r.Header(),
		r.Context(),
	)
}

func (r Request) UriPattern() string {
	return r.GetContext(ContextKeyUriPattern).(string)
}

func (r Request) Path() string {
	return r.r.URL.Path
}

func (r Request) PathParams() map[string]string {
	return r.GetContext(ContextKeyPathParams).(map[string]string)
}

func (r Request) QueryParams() map[string]string {
	return r.GetContext(ContextKeyQueryParams).(map[string]string)
}

func (r Request) Method() string {
	return r.r.Method
}

func (r Request) Header() http.Header {
	return r.r.Header
}

func (r Request) Body() ([]byte, error) {
	return ioutil.ReadAll(r.r.Body)
}

func (r Request) UnmarshalBody(holder interface{}) error {
	bodyStream, err := r.Body()
	if err != nil {
		return err
	}
	return json.Unmarshal(bodyStream, holder)
}

func (r Request) RemoteAddress() string {
	return r.r.RemoteAddr
}

func (r Request) GetContext(key string) interface{} {
	return r.c.GetContext(key)
}

func (r Request) RegisterContext(key string, value interface{}) {
	r.c.RegisterContext(key, value)
}

func (r Request) UnRegisterContext(key string) bool {
	return r.c.UnRegisterContext(key)
}

func (r Request) Context() RequestContext {
	return r.c
}
