package server

import (
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"net"
	"net/http"
)

type Engine func(listener net.Listener, server http.Handler) error

func NetEngine(listener net.Listener, server http.Handler) error {
	return http.Serve(listener, server)
}

func FastHTTPEngine(listener net.Listener, server http.Handler) error {
	return fasthttp.Serve(listener, fasthttpadaptor.NewFastHTTPHandler(server))
}
