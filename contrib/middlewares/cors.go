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
	ctx.Response().SetHeader(HeaderKeyAllowOrigin, "*")
	ctx.Response().SetHeader(HeaderKeyAllowHeaders, "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	ctx.Response().SetHeader(HeaderKeyVary, "Origin")
	if ctx.Request().Method() == http.MethodOptions {
		ctx.Report(nil)
		allowMethods := ctx.Request().MatchedService().SupportedMethodsForPattern(ctx.Request().UriPattern())
		allowMethodsHeaderValue := ""
		for _, method := range allowMethods {
			if method == "OPTIONS" {
				continue
			}
			allowMethodsHeaderValue += method + ", "
		}
		allowMethodsHeaderValue += "OPTIONS"
		ctx.Response().SetHeader(HeaderKeyAllowMethods, allowMethodsHeaderValue)
		ctx.Response().SetCode(http.StatusOK)
	}
}
