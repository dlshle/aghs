package middlewares

import (
	"github.com/dlshle/aghs/server"
	"net/http"
)

const (
	HeaderKeyAllowOrigin  = "Access-Control-Allow-Origin"
	HeaderKeyAllowMethods = "Access-Control-Allow-Methods"
	HeaderKeyAllowHeaders = "Access-Control-Allow-Headers"
	HeaderKeyVary         = "Vary"
)

func CORSAllowWildcardMiddleware(ctx server.MiddlewareContext) {
	ctx.Next()
	if ctx.Request().Method() == http.MethodOptions {
		ctx.Report(nil)
		ctx.Response().SetCode(http.StatusNoContent)
	}
	ctx.Response().SetHeader(HeaderKeyAllowMethods, "*")
	ctx.Response().SetHeader(HeaderKeyAllowOrigin, "*")
	ctx.Response().SetHeader(HeaderKeyAllowHeaders, "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	ctx.Response().SetHeader(HeaderKeyVary, "Origin")
}
