package server

import (
	"github.com/dlshle/gommon/errors"
	"github.com/dlshle/gommon/utils"
	"strings"
)

// chandle(request).Data() -> gets unmarshalled data
// chandle(request).QueryParam(x) -> string
// chandle(request).PathParam(y) -> string

type CHandle interface {
	Data() interface{}
	QueryParam(string) string
	PathParam(string) string
	Header(string) string
	Request() Request
}

type cHandle struct {
	data    interface{}
	request Request
}

func (h cHandle) Data() interface{} {
	return h.data
}

func (h cHandle) QueryParam(key string) string {
	return h.request.QueryParams()[key]
}

func (h cHandle) PathParam(key string) string {
	return h.request.QueryParams()[key]
}

func (h cHandle) Header(key string) string {
	return h.request.Header().Get(key)
}

func (h cHandle) Request() Request {
	return h.request
}

type cHandler struct {
	unmarshalFactory       func([]byte) (interface{}, error)
	isDataRequired         bool
	requiredPathParams     map[string]bool
	requiredQueryParams    map[string]bool
	requiredHeaderFields   map[string]bool
	onRequestHandle        func(CHandle) Response
	onErrorResponseFactory func(error) interface{}
}

func (h cHandler) HandleRequest(r Request) (Response, ServiceError) {
	var (
		data interface{}
		err  error
	)
	err = utils.ProcessWithErrors(func() error {
		data, err = h.getAndCheckData(r)
		return err
	}, func() error {
		return h.checkRequiredPathParams(r)
	}, func() error {
		return h.checkRequiredQueryParams(r)
	}, func() error {
		return h.checkRequiredHeaderFields(r)
	})
	if err != nil {
		if IsMissingRequiredFieldError(err) {
			return BadRequestResponse(h.handleError(err)), nil
		}
		return InternalServerErrorResponse(h.handleError(err)), nil
	}
	handle := cHandle{
		data:    data,
		request: r,
	}
	return h.onRequestHandle(handle), nil
}

func (h cHandler) checkRequiredHeaderFields(request Request) error {
	for key := range h.requiredHeaderFields {
		if request.Header().Get(key) == "" {
			return errors.Error("bad request: required header field " + key + " is missing")
		}
	}
	return nil
}

func (h cHandler) checkRequiredPathParams(request Request) error {
	pathParams := request.PathParams()
	for k := range h.requiredPathParams {
		if pathParams[k] == "" {
			return errors.Error("bad request: required path parameter " + k + " is missing")
		}
	}
	return nil
}

func (h cHandler) checkRequiredQueryParams(request Request) error {
	queryParams := request.QueryParams()
	for k := range h.requiredQueryParams {
		if queryParams[k] == "" {
			return errors.Error("bad request: required query parameter " + k + " is missing")
		}
	}
	return nil
}

func (h cHandler) getAndCheckData(request Request) (interface{}, error) {
	data, err := request.Body()
	if err != nil {
		return nil, err
	}
	if len(data) == 0 && h.isDataRequired {
		return nil, errors.Error("bad request: request body is missing")
	}
	return h.unmarshalFactory(data)
}

func (h cHandler) handleError(err error) interface{} {
	if h.onErrorResponseFactory != nil {
		return h.onErrorResponseFactory(err)
	}
	return err.Error()
}

type CHandlerBuilder interface {
	AddRequiredQueryParam(key string) CHandlerBuilder
	AddRequiredPathParam(key string) CHandlerBuilder
	AddRequiredHeaderField(key string) CHandlerBuilder
	RequireBody() CHandlerBuilder
	Unmarshaller(func([]byte) (interface{}, error)) CHandlerBuilder
	ErrorHandler(func(error) interface{}) CHandlerBuilder
	OnRequest(func(CHandle) Response) CHandlerBuilder
	Build() (cHandler, error)
}

type cHandlerBuilder struct {
	cHandlerRef *cHandler
}

func NewCHandlerBuilder() CHandlerBuilder {
	return &cHandlerBuilder{
		&cHandler{
			requiredPathParams:   make(map[string]bool),
			requiredQueryParams:  make(map[string]bool),
			requiredHeaderFields: make(map[string]bool),
		},
	}
}

func (b *cHandlerBuilder) AddRequiredHeaderField(key string) CHandlerBuilder {
	b.cHandlerRef.requiredHeaderFields[key] = true
	return b
}

func (b *cHandlerBuilder) AddRequiredQueryParam(key string) CHandlerBuilder {
	b.cHandlerRef.requiredQueryParams[key] = true
	return b
}

func (b *cHandlerBuilder) AddRequiredPathParam(key string) CHandlerBuilder {
	b.cHandlerRef.requiredPathParams[key] = true
	return b
}

func (b *cHandlerBuilder) RequireBody() CHandlerBuilder {
	b.cHandlerRef.isDataRequired = true
	return b
}

func (b *cHandlerBuilder) Unmarshaller(unmarshaller func([]byte) (interface{}, error)) CHandlerBuilder {
	b.cHandlerRef.unmarshalFactory = unmarshaller
	return b
}

func (b *cHandlerBuilder) ErrorHandler(callback func(error) interface{}) CHandlerBuilder {
	b.cHandlerRef.onErrorResponseFactory = callback
	return b
}

func (b *cHandlerBuilder) OnRequest(handler func(CHandle) Response) CHandlerBuilder {
	b.cHandlerRef.onRequestHandle = handler
	return b
}

func (b *cHandlerBuilder) Build() (h cHandler, e error) {
	if b.cHandlerRef.onRequestHandle == nil {
		e = errors.Error("no request handler set")
		return
	}
	if b.cHandlerRef.isDataRequired && b.cHandlerRef.unmarshalFactory == nil {
		e = errors.Error("no unmarshaller set when data is required")
		return
	}
	h = *(b.cHandlerRef)
	return
}

func IsMissingRequiredFieldError(err error) bool {
	return strings.HasPrefix(err.Error(), "bad request:")
}
