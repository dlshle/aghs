package server

import (
	"net"
	"net/http"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

type Engine func(listener net.Listener, server http.Handler) error

func NetEngine(listener net.Listener, server http.Handler) error {
	return http.Serve(listener, server)
}

func FastHTTPEngine(listener net.Listener, server http.Handler) error {
	fastHTTPServer := fasthttp.Server{
		Handler:               fasthttpadaptor.NewFastHTTPHandler(server),
		NoDefaultServerHeader: true,
	}
	return fastHTTPServer.Serve(listener)
}
