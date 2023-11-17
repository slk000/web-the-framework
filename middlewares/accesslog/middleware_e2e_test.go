//go:build e2e

package accesslog

import (
	"fmt"
	"github.com/slk000/web-the-framework"
	"testing"
)

func TestAccesslogMiddleware(t *testing.T) {
	// build the middleware
	accesslog := (&Builder{}).RegisterWriter(func(s string) {
		fmt.Println(s)
	}).Build()
	// config into server
	server := web.NewHTTPServer(web.GetHTTPMiddlewareConfigurer(accesslog))
	server.Get("/", func(ctx *web.Context) {
		ctx.ResponseContent = []byte("hi")
	})

	server.Start(":8001")
}
