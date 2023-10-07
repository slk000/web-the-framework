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
}

// NewHTTPServer constructs a http server
func NewHTTPServer() *HTTPServer {
	return &HTTPServer{
		router: newRouter(),
	}
}

type HTTPSServer struct {
}

// ServeHTTP serves an HTTP request: parses route and executes handler
func (h *HTTPServer) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	ctx := &Context{
		Req:  request,
		Resp: writer,
	}

	routeNode, pathParam := h.router.FindRoute(ctx.Req.Method, ctx.Req.URL.Path)
	if routeNode == nil || routeNode.handler == nil {
		http.NotFound(ctx.Resp, ctx.Req)
		return
	}
	ctx.Param = pathParam
	routeNode.handler(ctx)
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
