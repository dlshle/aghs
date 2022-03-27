package main

import (
	"fmt"
	"github.com/dlshle/aghs/server"
	"strconv"
)

func main() {
	counter := 0
	httpServer := server.New("0.0.0.0:1234")
	err := httpServer.RegisterService(NewStudentService())
	httpServer.Use(func(ctx server.MiddlewareContext) {
		counter++
		fmt.Println("request addr:", ctx.Request().RemoteAddress())
		defer func() {
			ctx.Response().SetHeader("hello", "world")
			ctx.Response().SetHeader("counter", strconv.Itoa(counter))
		}()
		if counter >= 2 {
			ctx.Report(server.NewServiceError(400, "badbad"))
			return
		}
		ctx.Next()
	})
	if err != nil {
		panic(err)
	}
	httpServer.Start()
}
