package server

import (
	"encoding/json"
	"net/http"
	"sync"
)

var responsePool sync.Pool = sync.Pool{New: func() any {
	return new(response)
}}

type Response interface {
	Code() int
	SetCode(code int)
	Payload() interface{}
	SetPayload(payload interface{})
	SetHeader(key string, value string)
	GetHeader(key string) (bool, string)
	IterateHeaders(cb func(k, v string))
	PayloadStream() (stream []byte, err error)
	ContentType() string
}

type response struct {
	code        int
	payload     interface{}
	contentType string
	header      map[string]string
}

func NewResponse(code int, payload interface{}) Response {
	return NewResponseWithContentType(code, payload, "application/json; charset=UTF-8")
}

func NewPlainTextResponse(code int, payload interface{}) Response {
	return NewResponseWithContentType(code, payload, "text/plain")
}

func NewResponseWithContentType(code int, payload interface{}, contentType string) Response {
	r := responsePool.Get().(*response)
	r.code = code
	r.payload = payload
	r.contentType = contentType
	r.header = make(map[string]string)
	return r
}

func (r *response) recycle() {
	r.code = -1
	r.payload = nil
	r.contentType = ""
	r.header = nil
	responsePool.Put(r)
}

func (r *response) Code() int {
	return r.code
}

func (r *response) SetCode(code int) {
	r.code = code
}

func (r *response) Payload() interface{} {
	return r.payload
}

func (r *response) SetPayload(payload interface{}) {
	r.payload = payload
}

func (r *response) SetHeader(key string, value string) {
	r.header[key] = value
}

func (r *response) GetHeader(key string) (bool, string) {
	value, exists := r.header[key]
	return exists, value
}

func (r *response) IterateHeaders(cb func(k, v string)) {
	for k, v := range r.header {
		cb(k, v)
	}
}

func (r *response) PayloadStream() (stream []byte, err error) {
	if r.payload == nil || r.code == http.StatusNoContent {
		return nil, nil
	}
	switch r.payload.(type) {
	case []byte:
		stream = r.payload.([]byte)
	case string:
		stream = []byte(r.payload.(string))
	default:
		stream, err = json.Marshal(r.Payload())
	}
	return
}

func (r *response) SetContentType(contentType string) {
	r.contentType = contentType
}

func (r *response) ContentType() string {
	if r.payload == nil || r.code == http.StatusNoContent {
		return ""
	}
	return r.contentType
}

func InternalServerErrorResponse(payload interface{}) Response {
	return NewResponse(http.StatusInternalServerError, payload)
}

func MethodNotAllowedResponse(payload interface{}) Response {
	return NewResponse(http.StatusMethodNotAllowed, payload)
}

func NotFoundResponse(payload interface{}) Response {
	return NewResponse(http.StatusNotFound, payload)
}

func BadRequestResponse(payload interface{}) Response {
	return NewResponse(http.StatusBadRequest, payload)
}

func ForbiddenResponse(payload interface{}) Response {
	return NewResponse(http.StatusForbidden, payload)
}
