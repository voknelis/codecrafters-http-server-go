package http

type RouteHandler = func(ResponseWriter, *Request) error

type Route struct {
	Method  string
	Pattern string
	Handler RouteHandler
}
