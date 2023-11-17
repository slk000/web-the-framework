package web

import (
	"net"
	"net/http"
)

// Ensure HTTPServer implements Server
var _ Server = &HTTPServer{}

// HandleFunc is the type of handler function
type HandleFunc func(ctx *Context)
type param map[string]string

// Server is an abstract type for any kind of server
type Server interface {
	http.Handler
	Start(addr string) error
	AddRoute(method string, path string, handleFunc HandleFunc)
}

// HTTPServer is a server handling  HTTP request
type HTTPServer struct {
	*router
	middlewares []Middleware // server-level middlewares
}

// HTTPServerConfigurer is a function that configures the http server
type HTTPServerConfigurer func(server *HTTPServer) *HTTPServer

// NewHTTPServer constructs a http server
func NewHTTPServer(configurers ...HTTPServerConfigurer) *HTTPServer {
	server := &HTTPServer{
		router: newRouter(),
	}
	for _, configurer := range configurers {
		configurer(server)
	}
	return server
}

// GetHTTPMiddlewareConfigurer builds a middleware configurer with provided middlewares
func GetHTTPMiddlewareConfigurer(middlewares ...Middleware) HTTPServerConfigurer {
	return func(server *HTTPServer) *HTTPServer {
		server.middlewares = middlewares // use provided middlewares to replace server.middlewares.

		return server
	}
}

type HTTPSServer struct {
}

// ServeHTTP serves an HTTP request: parses route and executes handler
func (h *HTTPServer) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	ctx := &Context{
		Req:  request,
		resp: writer,
	}

	// find a match router handler
	routeNode, pathParam := h.router.FindRoute(ctx.Req.Method, ctx.Req.URL.Path)
	if routeNode == nil || routeNode.handler == nil {
		http.NotFound(ctx.resp, ctx.Req)
		return
	}
	ctx.Param = pathParam

	// wrap the router handler with middleware handlers
	// TODO: should middlewares work when router not found?
	wrappedHandler := routeNode.handler
	for i := len(h.middlewares) - 1; i >= 0; i-- {
		wrappedHandler = h.middlewares[i](wrappedHandler)
	}

	wrappedHandler(ctx)
	// Sends final response
	ctx.resp.WriteHeader(ctx.StatusCode)
	ctx.resp.Write(ctx.ResponseContent)
}

// Start the HTTPServer
func (h *HTTPServer) Start(addr string) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	// here you are allowed to register your own actions after net.Listen and before http.Serve

	return http.Serve(l, h)
}

// Get request tool function
func (h *HTTPServer) Get(path string, handler HandleFunc) {
	h.AddRoute(http.MethodGet, path, handler)
}

// Post request tool function
func (h *HTTPServer) Post(path string, handler HandleFunc) {
	h.AddRoute(http.MethodPost, path, handler)
}
