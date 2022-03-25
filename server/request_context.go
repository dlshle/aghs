package server

type RequestContext map[string]interface{}

func (c RequestContext) GetContext(key string) interface{} {
	return c[key]
}

func (c RequestContext) RegisterContext(key string, value interface{}) {
	c[key] = value
}

func (c RequestContext) UnRegisterContext(key string) bool {
	if _, exists := c[key]; !exists {
		return false
	}
	delete(c, key)
	return true
}
