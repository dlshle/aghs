package middlewares

import (
	"net/http"

	"github.com/dlshle/aghs/server"
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
	ctx.Response().SetHeader(HeaderKeyAllowHeaders, "*")
	ctx.Response().SetHeader(HeaderKeyVary, "Origin")
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
	if ctx.Request().Method() == http.MethodOptions {
		ctx.Response().SetCode(http.StatusOK)
	}
}
