package server

import (
	"encoding/json"
	"github.com/dlshle/aghs/utils"
)

type Response struct {
	code        int
	payload     interface{}
	contentType string
	header      map[string]string
}

func NewResponse(code int, payload interface{}) *Response {
	return &Response{code, payload, "application/json; charset=UTF-8", make(map[string]string)}
}

func NewPlainTextResponse(code int, payload interface{}) *Response {
	return NewResponseWithContentType(code, payload, "text/plain")
}

func NewResponseWithContentType(code int, payload interface{}, contentType string) *Response {
	return &Response{code, payload, contentType, make(map[string]string)}
}

func (r *Response) Code() int {
	return r.code
}

func (r *Response) SetCode(code int) {
	r.code = code
}

func (r *Response) Payload() interface{} {
	return r.payload
}

func (r *Response) SetPayload(payload interface{}) {
	r.payload = payload
}

func (r *Response) SetHeader(key string, value string) {
	r.header[key] = value
}

func (r *Response) GetHeader(key string) (bool, string) {
	value, exists := r.header[key]
	return exists, value
}

func (r *Response) IterateHeaders(cb func(k, v string)) {
	for k, v := range r.header {
		cb(k, v)
	}
}

func (r *Response) PayloadStream() (stream []byte, err error) {
	switch r.payload.(type) {
	case []byte:
		stream = r.payload.([]byte)
		break
	case string:
		stream = []byte(r.payload.(string))
		break
	default:
		if transformed, ok := r.payload.(utils.Stringify); ok {
			stream = []byte(transformed.String())
		} else {
			stream, err = json.Marshal(r.Payload())
		}
	}
	return
}

func (r *Response) ContentType() string {
	return r.contentType
}
