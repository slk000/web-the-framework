//g o:build e2e

// By adding go:build e2e to the top of the file, these tests do not run with `go test`,
// but are invoked with `go test -tags e2e`
package web

import (
    "fmt"
    "net/http"
    "testing"
)

func TestServerE2e(t *testing.T) {
    h := NewHTTPServer()

    h.AddRoute(http.MethodGet, "/*", func(ctx *Context) {
        ctx.StatusCode = http.StatusOK
        ctx.ResponseContent = []byte("what's your path?")
    })
    h.AddRoute(http.MethodGet, "/user/*/home", func(ctx *Context) {
        ctx.StatusCode = http.StatusOK
        ctx.ResponseContent = []byte(fmt.Sprintf("hello /user/*/home (%s)\n", ctx.Req.URL.Path))
    })
    h.AddRoute(http.MethodGet, "/user/nobody/home", func(ctx *Context) {
        ctx.StatusCode = http.StatusOK
        ctx.ResponseContent = []byte("nobody lives here")
    })
    // trailing wildcard VS more specific route
    // only something like "/user/something/home" gets matched with "/user/*/home".
    // "/user/*/not_home" fallback to :/user/*".
    // we DO NOT support "user/*/not_home/something" fallback to "/user/*
    h.AddRoute(http.MethodGet, "/user/*", func(ctx *Context) {
        ctx.StatusCode = http.StatusOK
        ctx.ResponseContent = []byte(fmt.Sprintf("/user/* %s\n", ctx.Req.URL.Path))
    })

    h.AddRoute(http.MethodGet, "/index", func(ctx *Context) {
        ctx.StatusCode = http.StatusOK
        ctx.ResponseContent = []byte("hello /index")
    })

    // param
    h.AddRoute(http.MethodGet, "/index/:msg1/bruh/:msg2", func(ctx *Context) {
        ctx.StatusCode = http.StatusOK
        ctx.ResponseContent = []byte(fmt.Sprintf("hello /index/:msg1/bruh/:msg2 [%v]", *ctx.Param))
    })

    // trailing wildcard matches anything after it --> /a/(b/c/d/e/f...)
    h.AddRoute(http.MethodGet, "/a/*", func(ctx *Context) {
        ctx.StatusCode = http.StatusOK
        ctx.ResponseContent = []byte(fmt.Sprintf("trailing wildcard /a/* %s\n", ctx.Req.URL.Path))
    })
    // not a trailing wildcard
    h.AddRoute(http.MethodGet, "/aa/*/bb", func(ctx *Context) {
        ctx.StatusCode = http.StatusOK
        ctx.ResponseContent = []byte(fmt.Sprintf("/aa/*/bb %s\n", ctx.Req.URL.Path))
    })
    // /aa/*/cc/* (also match anything left)
    h.AddRoute(http.MethodGet, "/aa/*/cc/*", func(ctx *Context) {
        ctx.StatusCode = http.StatusOK
        ctx.ResponseContent = []byte(fmt.Sprintf("/aa/*/cc/* %s\n", ctx.Req.URL.Path))
    })

    // Regexp
    h.AddRoute(http.MethodGet, "/regexp/:key((\\d+))", func(ctx *Context) {
        ctx.StatusCode = http.StatusOK
        ctx.ResponseContent = []byte(fmt.Sprintf("Route: /regexp/:key((\\d+)) \nPath: %s", ctx.Req.URL.Path))
    })
    h.AddRoute(http.MethodGet, "/regexp/user/:role((.+)_.+)", func(ctx *Context) {
        ctx.StatusCode = http.StatusOK
        ctx.ResponseContent = []byte(fmt.Sprintf("Route: /regexp/:role((.+)_.+) \nPath: %s\nParams:%v", ctx.Req.URL.Path, ctx.Param))
    })
    h.AddRoute(http.MethodGet, "/regexp/multi1/:key1:key2:key3((\\d+)([a-z]+)(\\d+))", func(ctx *Context) {
        ctx.StatusCode = http.StatusOK
        ctx.ResponseContent = []byte(fmt.Sprintf("Route: /regexp/multi/:key1:key2:key3((\\d+)([a-z]+)(\\d+))\nPath: %s\nParams: %v", ctx.Req.URL.Path, ctx.Param))
    })
    h.AddRoute(http.MethodGet, "/regexp/multi2/:key2:key3(\\d+([a-z]+)(\\d+))", func(ctx *Context) {
        ctx.StatusCode = http.StatusOK
        ctx.ResponseContent = []byte(fmt.Sprintf("Route: /regexp/multi/:key1:key2:key3(\\d+([a-z]+)(\\d+))\nPath: %s\nParams: %v", ctx.Req.URL.Path, ctx.Param))
    })
    // Usage 1: delegate to http package
    // http.ListenAndServe("8081", h)
    // http.ListenAndServeTLS("443", "", "", h)

    // Usage 2: self managed
    h.Start(":8001")
}
