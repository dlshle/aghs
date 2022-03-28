package middlewares

import (
	"github.com/dlshle/aghs/server"
	"net/http"
)

const (
	HeaderKeyAllowOrigin  = "Access-Control-Allow-Origin"
	HeaderKeyAllowMethods = "Access-Control-Allow-Methods"
)

func CORSAllowWildcardMiddleware(ctx server.MiddlewareContext) {
	ctx.Next()
	if ctx.Request().Method() == http.MethodOptions {
		ctx.Report(nil)
		ctx.Response().SetCode(http.StatusOK)
	}
	ctx.Response().SetHeader(HeaderKeyAllowMethods, "*")
	ctx.Response().SetHeader(HeaderKeyAllowOrigin, "*")
}
