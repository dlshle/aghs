package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/dlshle/aghs/utils"
)

type ServiceError interface {
	Error() string
	Code() int
	ContentType() string
	AttachContext(ctx context.Context)
}

type serviceError struct {
	code int    // code should correspond to an HTTP error code
	msg  string // this will be the payload for response
	ctx  context.Context
}

func NewServiceErrorWithCode(code int, msg string) *serviceError {
	if code < 400 || code > 600 {
		code = 500
	}
	msg = strings.TrimSpace(msg)
	return &serviceError{code, msg, nil}
}

func NewServiceError(msg string) *serviceError {
	return NewServiceErrorWithCode(http.StatusInternalServerError, msg)
}

func (e *serviceError) Error() string {
	if e.ctx != nil {
		return e.errorWithContext()
	}
	return e.msg
}

func (e *serviceError) errorWithContext() string {
	var builder strings.Builder
	builder.WriteByte('{')
	builder.WriteString(fmt.Sprintf(`"message":"%s"`, utils.EncodeString(e.msg)))
	builder.WriteByte('}')
	return builder.String()
}

func (e *serviceError) ContentType() string {
	if len(e.msg) > 0 && e.msg[0] == '{' {
		return "application/json; charset=UTF-8"
	}
	return "text/plain"
}

func (e *serviceError) Code() int {
	return e.code
}

func (e *serviceError) AttachContext(ctx context.Context) {
	e.ctx = ctx
}

func MethodNotAllowedError(msg string) *serviceError {
	return NewServiceErrorWithCode(http.StatusMethodNotAllowed, msg)
}

func NotFoundError(msg string) *serviceError {
	return NewServiceErrorWithCode(http.StatusNotFound, msg)
}

func BadRequestError(msg string) *serviceError {
	return NewServiceErrorWithCode(http.StatusBadRequest, msg)
}

func InternalError(msg string) *serviceError {
	return NewServiceErrorWithCode(http.StatusInternalServerError, msg)
}

func ForbiddenError(msg string) *serviceError {
	return NewServiceErrorWithCode(http.StatusForbidden, msg)
}
