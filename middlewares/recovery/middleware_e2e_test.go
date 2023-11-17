//go:build e2e

package recovery

import (
	"fmt"
	"github.com/slk000/web-the-framework"
	"net/http"
	"testing"
)

func TestRecoveryMiddlewareE2e(t *testing.T) {
	middleware := (&Builder{}).SetResp(http.StatusInternalServerError, "Internal Server Error!!!").Build()
	server := web.NewHTTPServer(web.GetHTTPMiddlewareConfigurer(middleware))
	server.Get("/", func(ctx *web.Context) {
		ctx.StatusCode = http.StatusOK
		ctx.ResponseContent = []byte("Hi!")
		panic("oh shoot")
	})
	server.Start(":8001")
}

func TestRecoveryMiddlewareWithCustomMsgE2e(t *testing.T) {
	builder := Builder{
		StatusCode: 418,
		ContentFunc: func(err any) []byte {
			return []byte(fmt.Sprintf("I am a teapot with err: %v", err))
		},
	}
	middleware := builder.Build()
	server := web.NewHTTPServer(web.GetHTTPMiddlewareConfigurer(middleware))
	server.Get("/", func(ctx *web.Context) {
		ctx.StatusCode = http.StatusOK
		ctx.ResponseContent = []byte("Hi!")
		panic("oh shoot")
	})
	server.Start(":8001")
}
