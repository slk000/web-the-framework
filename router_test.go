package web

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test add correct static paths
func TestRouter_AddRouteStaticPath(t *testing.T) {
	// The paths
	testRoutes := []struct {
		method string
		path   string
	}{
		{
			method: http.MethodGet,
			path:   "/",
		},
		{
			method: http.MethodGet,
			path:   "/user",
		},
		{
			method: http.MethodGet,
			path:   "/user/home",
		},
		{
			method: http.MethodGet,
			path:   "/order/detail",
		},
	}
	var mockHandler = func(ctx *Context) {}

	// Construct the router
	r := newRouter()
	for _, route := range testRoutes {
		r.AddRoute(route.method, route.path, mockHandler)
	}

	/*
					*"/"
				/		  \
		*"user"			"order"
		   |			   |
		*"home"			*"detail"
	*/

	// The expected router structure
	wantRouter := &router{
		trees: map[string]*node{
			http.MethodGet: &node{
				path:    "/",
				handler: mockHandler,
				children: map[string]*node{
					"user": &node{
						path:    "user",
						handler: mockHandler,
						children: map[string]*node{
							"home": &node{
								path:    "home",
								handler: mockHandler,
							},
						},
					},
					"order": &node{
						path: "order",
						children: map[string]*node{
							"detail": &node{
								path:    "detail",
								handler: mockHandler,
							},
						},
					},
				},
			},
		},
	}

	// Test
	msg, ok := wantRouter.equal(r)
	assert.True(t, ok, msg)
}

// Test add incorrect static paths
func TestRouter_AddRoute2(t *testing.T) {
	var mockHandler = func(ctx *Context) {}
	r := newRouter()
	assert.Panicsf(t, func() {
		r.AddRoute(http.MethodGet, "", mockHandler)
	}, "empty path")
	assert.Panicsf(t, func() {
		r.AddRoute(http.MethodGet, "/a/", mockHandler)
	}, "end with /")
	assert.Panicsf(t, func() {
		r.AddRoute(http.MethodGet, "d/d", mockHandler)
	}, "not start with /")
	assert.Panicsf(t, func() {
		r.AddRoute(http.MethodGet, "/d//d", mockHandler)
	}, "continuous /")

	r.AddRoute(http.MethodGet, "/", mockHandler)
	assert.Panicsf(t, func() {
		r.AddRoute(http.MethodGet, "/", mockHandler)
	}, "duplicate root node")
	r.AddRoute(http.MethodGet, "/a", mockHandler)
	assert.Panicsf(t, func() {
		r.AddRoute(http.MethodGet, "/a", mockHandler)
	}, "duplicate node")
}

// Test find route node by path
func TestRouter_FindRoute(t *testing.T) {
	// The paths to construct router
	testRoutes := []struct {
		method string
		path   string
	}{
		{
			method: http.MethodPost,
			path:   "/",
		},
		{
			method: http.MethodGet,
			path:   "/user",
		},
		{
			method: http.MethodGet,
			path:   "/user/home",
		},
		{
			method: http.MethodGet,
			path:   "/order/detail",
		},
	}

	var mockHandler = func(ctx *Context) {}

	r := newRouter()
	for _, route := range testRoutes {
		r.AddRoute(route.method, route.path, mockHandler)
	}

	// Cases of path
	testCases := []struct {
		caseName   string
		method     string
		fullPath   string
		wantedNode *node // Expected node
	}{
		{
			caseName: "find /user",
			method:   http.MethodGet,
			fullPath: "/user",
			wantedNode: &node{
				path: "user",
				children: map[string]*node{
					"home": &node{path: "home", handler: mockHandler},
				},
				handler: mockHandler,
			},
		},
		{
			caseName:   "find non-exist path /user/no",
			method:     http.MethodGet,
			fullPath:   "/user/no",
			wantedNode: nil, // FindRoute should not find a node
		},
		{
			caseName:   "find non-exist path /user/home/no",
			method:     http.MethodGet,
			fullPath:   "/user/home/no",
			wantedNode: nil,
		},
		{
			caseName: "root",
			method:   http.MethodPost,
			fullPath: "/",
			wantedNode: &node{
				path:    "/",
				handler: mockHandler,
			},
		},
	}

	// run sub-testcases
	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			foundNode := r.FindRoute(tc.method, tc.fullPath)
			if tc.wantedNode == nil {
				assert.Nil(t, foundNode)
			} else {
				_, ok := tc.wantedNode.equal(foundNode)
				assert.True(t, ok)
			}
		})
	}
}
func (r *router) equal(y *router) (string, bool) {
	// 对森林中的每一个路由树进行比较
	for method, tree := range r.trees {
		dst, ok := y.trees[method]
		if !ok {
			return fmt.Sprint("找不到对应的http method"), false
		}
		// 递归比较这个树中的各节点
		msg, equal := tree.equal(dst)
		if !equal {
			return msg, false
		}
	}
	return "", true

}
func (n *node) equal(that *node) (string, bool) {
	if n.path != that.path {
		return fmt.Sprint("节点路径不匹配"), false
	}
	if len(n.children) != len(that.children) {
		return fmt.Sprint("子节点数量不匹配"), false
	}

	// 由于func()不可比，通过反射来比较handler
	nHandler := reflect.ValueOf(n.handler)
	yHandler := reflect.ValueOf(that.handler)
	if nHandler != yHandler {

		return fmt.Sprint("handler 不相等"), false
	}

	for path, c := range n.children {
		dst, ok := that.children[path]
		if !ok {

			return fmt.Sprintf("子节点 %s 不存在", path), false
		}
		msg, ok := c.equal(dst)
		if !ok {
			return msg, false
		}

	}
	return "", true
}
