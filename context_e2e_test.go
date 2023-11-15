//go:build e2e

package web

import (
	"fmt"
	"net/http"
	"testing"
)

type data struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func TestContextE2e(t *testing.T) {
	h := NewHTTPServer()
	h.AddRoute(http.MethodPost, "/json", func(ctx *Context) {
		d := &data{0, "000"}
		err := ctx.BindJSON(d)
		if err != nil {
			ctx.Resp.WriteHeader(http.StatusBadRequest)
			ctx.Resp.Write([]byte("bad request format"))
			return
		}
		ctx.Resp.WriteHeader(http.StatusOK)
		ctx.Resp.Write([]byte(fmt.Sprintf("%v", d)))
	})
	h.Start(":8001")

}
