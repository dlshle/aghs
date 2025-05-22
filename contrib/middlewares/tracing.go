package middlewares

import (
	"github.com/dlshle/aghs/server"
	"github.com/dlshle/aghs/utils"
)

const (
	CtxKeyTraceID = "trace_id"
)

func TracingMiddleware(ctx server.MiddlewareContext) {
	ctx.Request().RegisterContext(CtxKeyTraceID, utils.GenerateID())
	ctx.Next()
}
