package server

type RequestHandler func(r Request) (Response, ServiceError)
