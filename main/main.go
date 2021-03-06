package main

import (
	"github.com/dlshle/aghs/contrib/middlewares"
	"github.com/dlshle/aghs/server"
	"strconv"
	"sync/atomic"
)

func main() {
	var requestCounter uint32 = 0
	httpServer, err := server.NewBuilder().
		Engine(server.FastHTTPEngine).
		Address("0.0.0.0:1234").
		WithService(NewStudentService()).
		WithMiddleware(middlewares.CORSAllowWildcardMiddleware).
		WithMiddleware(func(ctx server.MiddlewareContext) {
			atomic.AddUint32(&requestCounter, 1)
			defer func() {
				ctx.Response().SetHeader("total-request-count", strconv.Itoa(int(atomic.LoadUint32(&requestCounter))))
			}()
			ctx.Next()
		}).Build()
	if err != nil {
		panic(err)
	}
	httpServer.Start()
}
