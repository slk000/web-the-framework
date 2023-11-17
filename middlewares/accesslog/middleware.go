package accesslog

import (
	"fmt"
	"github.com/slk000/web-the-framework"
)

// Builder holds the build config of accesslog middleware
type Builder struct {
	writer func(string)
}

// Fields in the log
type accessLog struct {
	Host   string
	Method string
	Path   string
	Router string // matched router name
}

// RegisterWriter to the accesslog
func (b *Builder) RegisterWriter(writer func(string)) *Builder {
	b.writer = writer
	return b
}

// Build generates an accesslog middleware instance
func (m *Builder) Build() web.Middleware {
	return func(next web.HandleFunc) web.HandleFunc { // return the built accesslog middleware
		return func(ctx *web.Context) { // in the middleware, return the handler to deal with request
			defer func() {
				// Run after next(ctx) to get matched router name, and avoid panic in next(ctx)
				if m.writer == nil {
					panic("accesslog writer not registered.")
				}
				l := &accessLog{
					Host:   ctx.Req.Host,
					Method: ctx.Req.Method,
					Path:   ctx.Req.URL.Path,
					Router: "--",
				}
				m.writer(fmt.Sprintf("%v", l))
			}()
			next(ctx)
		}

	}
}
