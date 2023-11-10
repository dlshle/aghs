package health

import "github.com/dlshle/aghs/server"

func NewHealthCheckService(path string) server.Service {
	svc, _ := server.NewServiceBuilder().
		Id("health-check").
		WithRouteHandlers(server.PathHandlerBuilder(path).Get(func(r server.Request) (server.Response, server.ServiceError) {
			return server.NewResponse(200, "ok"), nil
		})).
		Build()
	return svc
}
