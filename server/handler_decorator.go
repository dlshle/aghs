package server

func DecorateBefore(preProcessor PreProcessor, handler RequestHandler) RequestHandler {
	return func(r Request) (Response, ServiceError) {
		processed, err := preProcessor(r)
		if err != nil {
			return nil, err
		}
		return handler(processed)
	}
}

func DecorateAfter(afterProcessor PostProcessor, handler RequestHandler) RequestHandler {
	return func(r Request) (Response, ServiceError) {
		resp, err := handler(r)
		return afterProcessor(resp, err)
	}
}
