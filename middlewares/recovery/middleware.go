package recovery

import (
	"github.com/slk000/web-the-framework"
)

// Builder holds the build config of recovery middleware
type Builder struct {
	StatusCode  int                  // http code to be sent
	ContentFunc func(err any) []byte // function that generates content to be sent. err is captured panic
}

// SetResp sets http code and simple content to be sent in the middleware
func (b *Builder) SetResp(code int, content string) *Builder {
	b.StatusCode = code
	b.ContentFunc = func(err any) []byte {
		return []byte(content)
	}
	return b
}

// Build generate a recovery middleware instance
func (b *Builder) Build() web.Middleware {
	return func(next web.HandleFunc) web.HandleFunc {
		return func(ctx *web.Context) {
			defer func() {
				if err := recover(); err != nil {
					ctx.StatusCode = b.StatusCode
					ctx.ResponseContent = b.ContentFunc(err)
				}
			}()
			next(ctx)
		}
	}
}
