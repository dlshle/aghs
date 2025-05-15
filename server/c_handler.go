package server

import (
	"io"
	"strings"

	"github.com/dlshle/gommon/errors"
	"github.com/dlshle/gommon/utils"
)

// chandle(request).Data() -> gets unmarshalled data
// chandle(request).QueryParam(x) -> string
// chandle(request).PathParam(y) -> string

type CHandle[T any] interface {
	Data() T
	Body() []byte
	QueryParam(string) string
	PathParam(string) string
	Header(string) string
	Request() Request
	FormFile(key string, maxSize int64) (io.ReadCloser, error)
	MultipartFormValue(key string, maxSize int64) ([]string, error)
	FormValue(key string) (string, error)
}

type cHandle[T any] struct {
	data    T
	request Request
}

func (h cHandle[T]) Data() T {
	return h.data
}

func (h cHandle[T]) Body() []byte {
	body, _ := h.request.Body()
	return body
}

func (h cHandle[T]) QueryParam(key string) string {
	return h.request.QueryParams()[key]
}

func (h cHandle[T]) PathParam(key string) string {
	return h.request.PathParams()[key]
}

func (h cHandle[T]) Header(key string) string {
	return h.request.Header().Get(key)
}

func (h cHandle[T]) Request() Request {
	return h.request
}

func (h cHandle[T]) FormFile(key string, maxSize int64) (io.ReadCloser, error) {
	return h.request.FormFile(key, maxSize)
}

func (h cHandle[T]) MultipartFormValue(key string, maxSize int64) ([]string, error) {
	return h.request.MultipartFormValue(key, maxSize)
}

func (h cHandle[T]) FormValue(key string) (string, error) {
	return h.request.FormValue(key)
}

type cHandler[T any] struct {
	unmarshalFactory       func([]byte) (T, error)
	isDataRequired         bool
	isFormDataRequired     bool
	requiredPathParams     map[string]bool
	requiredQueryParams    map[string]bool
	requiredHeaderFields   map[string]bool
	onRequestHandle        func(CHandle[T]) Response
	onErrorResponseFactory func(error) interface{}
}

func (h cHandler[T]) HandleRequest(r Request) (Response, ServiceError) {
	var (
		data T
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
	handle := cHandle[T]{
		data:    data,
		request: r,
	}
	return h.onRequestHandle(handle), nil
}

func (h cHandler[T]) checkRequiredHeaderFields(request Request) error {
	for key := range h.requiredHeaderFields {
		if request.Header().Get(key) == "" {
			return errors.Error("bad request: required header field " + key + " is missing")
		}
	}
	return nil
}

func (h cHandler[T]) checkRequiredPathParams(request Request) error {
	pathParams := request.PathParams()
	for k := range h.requiredPathParams {
		if pathParams[k] == "" {
			return errors.Error("bad request: required path parameter " + k + " is missing")
		}
	}
	return nil
}

func (h cHandler[T]) checkRequiredQueryParams(request Request) error {
	queryParams := request.QueryParams()
	for k := range h.requiredQueryParams {
		if queryParams[k] == "" {
			return errors.Error("bad request: required query parameter " + k + " is missing")
		}
	}
	return nil
}

func (h cHandler[T]) getAndCheckData(request Request) (T, error) {
	var zeroVal T
	if !h.isDataRequired {
		return zeroVal, nil
	}
	data, err := request.Body()
	if err != nil {
		return zeroVal, err
	}
	if len(data) == 0 && h.isDataRequired {
		return zeroVal, errors.Error("bad request: request body is missing")
	}
	if h.unmarshalFactory != nil {
		return h.unmarshalFactory(data)
	}
	return zeroVal, nil
}

func (h cHandler[T]) handleError(err error) interface{} {
	if h.onErrorResponseFactory != nil {
		return h.onErrorResponseFactory(err)
	}
	return err.Error()
}

type CHandlerBuilder[T any] interface {
	AddRequiredQueryParam(key string) CHandlerBuilder[T]
	AddRequiredPathParam(key string) CHandlerBuilder[T]
	AddRequiredHeaderField(key string) CHandlerBuilder[T]
	RequireBody() CHandlerBuilder[T]
	RequireFormData() CHandlerBuilder[T]
	Unmarshaller(func([]byte) (T, error)) CHandlerBuilder[T]
	UseDefaultUnmarshaller() CHandlerBuilder[T]
	ErrorHandler(func(error) interface{}) CHandlerBuilder[T]
	OnRequest(func(CHandle[T]) Response) CHandlerBuilder[T]
	Build() (cHandler[T], error)
	MustBuild() cHandler[T]
}

type cHandlerBuilder[T any] struct {
	cHandlerRef *cHandler[T]
}

func NewCHandlerBuilder[T any]() CHandlerBuilder[T] {
	return &cHandlerBuilder[T]{
		&cHandler[T]{
			requiredPathParams:   make(map[string]bool),
			requiredQueryParams:  make(map[string]bool),
			requiredHeaderFields: make(map[string]bool),
		},
	}
}

func (b *cHandlerBuilder[T]) AddRequiredHeaderField(key string) CHandlerBuilder[T] {
	b.cHandlerRef.requiredHeaderFields[key] = true
	return b
}

func (b *cHandlerBuilder[T]) AddRequiredQueryParam(key string) CHandlerBuilder[T] {
	b.cHandlerRef.requiredQueryParams[key] = true
	return b
}

func (b *cHandlerBuilder[T]) AddRequiredPathParam(key string) CHandlerBuilder[T] {
	b.cHandlerRef.requiredPathParams[key] = true
	return b
}

func (b *cHandlerBuilder[T]) RequireBody() CHandlerBuilder[T] {
	b.cHandlerRef.isDataRequired = true
	return b
}

func (b *cHandlerBuilder[T]) RequireFormData() CHandlerBuilder[T] {
	b.cHandlerRef.isFormDataRequired = true
	return b
}

func (b *cHandlerBuilder[T]) UseDefaultUnmarshaller() CHandlerBuilder[T] {
	b.cHandlerRef.unmarshalFactory = b.defaultUnmarshalFactory
	return b
}

func (b *cHandlerBuilder[T]) defaultUnmarshalFactory(data []byte) (T, error) {
	return utils.UnmarshalJSONEntity[T](data)
}

func (b *cHandlerBuilder[T]) Unmarshaller(unmarshaller func([]byte) (T, error)) CHandlerBuilder[T] {
	b.cHandlerRef.unmarshalFactory = unmarshaller
	return b
}

func (b *cHandlerBuilder[T]) ErrorHandler(callback func(error) interface{}) CHandlerBuilder[T] {
	b.cHandlerRef.onErrorResponseFactory = callback
	return b
}

func (b *cHandlerBuilder[T]) OnRequest(handler func(CHandle[T]) Response) CHandlerBuilder[T] {
	b.cHandlerRef.onRequestHandle = handler
	return b
}

func (b *cHandlerBuilder[T]) Build() (h cHandler[T], e error) {
	if b.cHandlerRef.onRequestHandle == nil {
		e = errors.Error("no request handler set")
		return
	}
	// we can have data w/out unmarshaller
	/*
		if b.cHandlerRef.isDataRequired && b.cHandlerRef.unmarshalFactory == nil {
			e = errors.Error("no unmarshaller set when data is required")
			return
		}
	*/
	h = *(b.cHandlerRef)
	return
}

func (b *cHandlerBuilder[T]) MustBuild() cHandler[T] {
	handler, err := b.Build()
	if err != nil {
		panic(err)
	}
	return handler
}

func DefaultBytesDataUnmarshaller(data []byte) ([]byte, error) {
	return data, nil
}

func IsMissingRequiredFieldError(err error) bool {
	return strings.HasPrefix(err.Error(), "bad request:")
}
