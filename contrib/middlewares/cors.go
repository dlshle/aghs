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
	// if no origin is present in the header, skip
	origin := ctx.Request().Header().Get("Origin")
	if origin == "" {
		return
	}

	// allow all for now
	ctx.Response().SetHeader(HeaderKeyAllowOrigin, origin)
	ctx.Response().SetHeader(HeaderKeyVary, "Origin")

	if ctx.Request().Method() == http.MethodOptions {
		// allow headers
		requestedHdrs := ctx.Request().Header().Get("Access-Control-Request-Headers")
		ctx.Response().SetHeader(HeaderKeyAllowHeaders, requestedHdrs)

		// allow methods
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
		ctx.Report(nil)
		ctx.Response().SetCode(http.StatusNoContent)
	}
}
