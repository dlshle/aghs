package server

type PreProcessor func(r Request) (Request, ServiceError)
type PostProcessor func(r *Response, err ServiceError) (*Response, ServiceError)
