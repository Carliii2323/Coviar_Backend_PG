package router

import (
	"context"
	"net/http"
	"regexp"
	"strings"
)

type Router struct {
	routes      []*Route
	middlewares []Middleware
	notFound    http.HandlerFunc
}

type Route struct {
	method      string
	pattern     *regexp.Regexp
	handler     http.HandlerFunc
	middlewares []Middleware
	paramNames  []string
}

type Middleware func(http.Handler) http.Handler

type contextKey string

const ParamsKey contextKey = "router_params"

func New() *Router {
	return &Router{
		routes: make([]*Route, 0),
		notFound: func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":"endpoint no encontrado"}`))
		},
	}
}

func (rt *Router) Use(middlewares ...Middleware) {
	rt.middlewares = append(rt.middlewares, middlewares...)
}

func (rt *Router) GET(pattern string, handler http.HandlerFunc) {
	rt.addRoute(http.MethodGet, pattern, handler)
}

func (rt *Router) POST(pattern string, handler http.HandlerFunc) {
	rt.addRoute(http.MethodPost, pattern, handler)
}

func (rt *Router) PUT(pattern string, handler http.HandlerFunc) {
	rt.addRoute(http.MethodPut, pattern, handler)
}

func (rt *Router) PATCH(pattern string, handler http.HandlerFunc) {
	rt.addRoute(http.MethodPatch, pattern, handler)
}

func (rt *Router) DELETE(pattern string, handler http.HandlerFunc) {
	rt.addRoute(http.MethodDelete, pattern, handler)
}

func (rt *Router) addRoute(method, pattern string, handler http.HandlerFunc) {
	regex, paramNames := compilePattern(pattern)
	route := &Route{
		method:      method,
		pattern:     regex,
		handler:     handler,
		paramNames:  paramNames,
		middlewares: make([]Middleware, len(rt.middlewares)),
	}
	copy(route.middlewares, rt.middlewares)
	rt.routes = append(rt.routes, route)
}

func (rt *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, route := range rt.routes {
		if route.method != r.Method {
			continue
		}

		matches := route.pattern.FindStringSubmatch(r.URL.Path)
		if matches == nil {
			continue
		}

		// Extraer parámetros
		if len(route.paramNames) > 0 {
			params := make(map[string]string)
			for i, name := range route.paramNames {
				params[name] = matches[i+1]
			}
			ctx := context.WithValue(r.Context(), ParamsKey, params)
			r = r.WithContext(ctx)
		}

		// Aplicar middlewares y ejecutar handler
		handler := route.handler
		for i := len(route.middlewares) - 1; i >= 0; i-- {
			handler = (route.middlewares[i])(http.HandlerFunc(handler)).ServeHTTP
		}
		handler(w, r)
		return
	}

	rt.notFound(w, r)
}

func compilePattern(pattern string) (*regexp.Regexp, []string) {
	paramNames := []string{}
	parts := strings.Split(pattern, "/")

	for i, part := range parts {
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			paramName := strings.TrimPrefix(strings.TrimSuffix(part, "}"), "{")
			paramNames = append(paramNames, paramName)
			parts[i] = "([^/]+)"
		}
	}

	regexPattern := "^" + strings.Join(parts, "/") + "$"
	regex := regexp.MustCompile(regexPattern)

	return regex, paramNames
}

// GetParam obtiene un parámetro de la ruta desde el contexto
func GetParam(r *http.Request, paramName string) string {
	if params, ok := r.Context().Value(ParamsKey).(map[string]string); ok {
		return params[paramName]
	}
	return ""
}
