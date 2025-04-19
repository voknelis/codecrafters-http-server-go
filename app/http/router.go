package http

import (
	"fmt"
	"strings"
)

type Router struct {
	routes map[string]Route
}

type RouteParams = map[string]string

func (r *Router) AddRoute(method, pattern string, handler RouteHandler) {
	route := Route{
		Method:  method,
		Pattern: pattern,
		Handler: handler,
	}
	key := r.getRouteKey(route.Method, route.Pattern)
	r.routes[key] = route
}

func (r *Router) getRouteKey(method, pattern string) string {
	return fmt.Sprintf("%s %s", method, pattern)
}

func (r *Router) parseRouterKey(key string) (string, string) {
	parts := strings.SplitN(key, " ", 2)
	return parts[0], parts[1]
}

func (r *Router) MatchRoute(method, pattern string) (RouteHandler, string) {
	for key, route := range r.routes {
		routeMethod, _ := r.parseRouterKey(key)
		if routeMethod != "*" && routeMethod != method {
			continue
		}

		ok := r.checkRouteMatch(pattern, route)
		if ok {
			return route.Handler, route.Pattern
		}
	}

	return nil, ""
}

func (r *Router) checkRouteMatch(targetPattent string, route Route) bool {
	targetSegments := strings.Split(route.Pattern, "/")
	routeSegments := strings.Split(targetPattent, "/")

	if len(targetSegments) != len(routeSegments) {
		return false
	}

	for i, targetSerment := range routeSegments {
		routeSegment := targetSegments[i]

		// check if template variable
		if strings.HasPrefix(routeSegment, "{") && strings.HasSuffix(routeSegment, "}") {
			continue
		}

		// non-matching segments
		if targetSerment != routeSegment {
			return false
		}
	}

	return true
}

func NewRouter() *Router {
	return &Router{
		routes: map[string]Route{},
	}
}
