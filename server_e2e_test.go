//go:build e2e

// By adding go:build e2e to the top of the file, these tests do not run with `go test`,
// but are invoked with `go test -tags e2e`
package web

import (
	"net/http"
	"testing"
)

func TestServer(t *testing.T) {
	h := NewHTTPServer()

	h.AddRoute(http.MethodGet, "/user", func(ctx *Context) {
		ctx.Resp.WriteHeader(http.StatusOK)
		ctx.Resp.Write([]byte("hello user"))
	})
	h.AddRoute(http.MethodGet, "/user/home", func(ctx *Context) {
		ctx.Resp.WriteHeader(http.StatusOK)
		ctx.Resp.Write([]byte("hello /user/home"))
	})
	h.AddRoute(http.MethodGet, "/index", func(ctx *Context) {
		ctx.Resp.WriteHeader(http.StatusOK)
		ctx.Resp.Write([]byte("hello /index"))
	})

	// Usage 1: delegate to http package
	// http.ListenAndServe("8081", h)
	// http.ListenAndServeTLS("443", "", "", h)

	// Usage 2: self managed
	h.Start(":8001")
}
