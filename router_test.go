package web

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test add correct static paths
func TestRouter_AddRouteStaticCorrect(t *testing.T) {
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
	ok, err := wantRouter.equal(r)
	assert.True(t, ok, err)
}

// Test add incorrect static paths
func TestRouter_AddRouteStaticIncorrect(t *testing.T) {
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

// Test add correct wildcard paths
func TestRouter_AddRouteWildcardCorrect(t *testing.T) {
	// The paths
	testRoutes := []struct {
		method string
		path   string
	}{
		{
			method: http.MethodGet,
			path:   "/user",
		},
		{
			method: http.MethodGet,
			path:   "/user/*",
		},
		{
			method: http.MethodGet,
			path:   "/user/*/home",
		},
		{
			method: http.MethodGet,
			path:   "/user/nobody/home",
		},
		{
			method: http.MethodGet,
			path:   "/*",
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
				*"user"		     *"*"
				 /  \
			 "nobody" "*"
			   /       \
		*"home(1)"   *"home(2)"


	*/

	// The expected router structure
	home1 := &node{
		path:    "home",
		handler: mockHandler,
	}
	home2 := &node{
		path:    "home",
		handler: mockHandler,
	}
	nobody := &node{
		path:     "nobody",
		children: map[string]*node{"home": home1},
	}
	wildcardOfUser := &node{
		children: map[string]*node{"home": home2},
	}
	user := &node{
		path:     "user",
		handler:  mockHandler,
		children: map[string]*node{"nobody": nobody},
		wildcard: wildcardOfUser,
	}
	wildcardOfRoot := &node{
		handler: mockHandler,
	}
	root := &node{
		path:     "/",
		children: map[string]*node{"user": user},
		wildcard: wildcardOfRoot,
	}

	wantRouter := &router{
		trees: map[string]*node{
			http.MethodGet: root,
		},
	}

	// Test
	ok, err := wantRouter.equal(r)
	assert.True(t, ok, err)
}

// Test find route node by static path
func TestRouter_FindRouteStatic(t *testing.T) {
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
				ok, err := tc.wantedNode.equal(foundNode)
				assert.True(t, ok, err)
			}
		})
	}
}

// Test find route node by wildcard path
func TestRouter_FindRouteWildcard(t *testing.T) {
	// The paths to construct router
	testRoutes := []struct {
		method string
		path   string
	}{
		{
			method: http.MethodGet,
			path:   "/*",
		},
		{
			method: http.MethodGet,
			path:   "/user/*/home",
		},
		{
			method: http.MethodGet,
			path:   "/user/nobody/home",
		},
		{
			method: http.MethodGet,
			path:   "/",
		},
		{
			method: http.MethodPost,
			path:   "/*", // no child other than wildcard
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
			caseName: "find /*",
			method:   http.MethodGet,
			fullPath: "/*",
			wantedNode: &node{
				path:    "*",
				handler: mockHandler,
			},
		},
		{
			caseName: "find /user/nobody/home",
			method:   http.MethodGet,
			fullPath: "/user/nobody/home",
			wantedNode: &node{
				path:    "home",
				handler: mockHandler,
			},
		},
		{
			caseName: "find /user/somebody/home",
			method:   http.MethodGet,
			fullPath: "/user/somebody/home",
			wantedNode: &node{
				path:    "home",
				handler: mockHandler,
			},
		},
		{
			caseName: "find non-exist /user/somebody/homo",
			method:   http.MethodGet,
			fullPath: "/user/somebody/homo",
		},
		{
			caseName: "a node has no child but a wildcard",
			method:   http.MethodPost,
			fullPath: "/bruh",
			wantedNode: &node{
				path:    "*",
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
				ok, err := tc.wantedNode.equal(foundNode)
				assert.True(t, ok, err)
			}
		})
	}
}

// Compare two routers.
func (r *router) equal(other *router) (bool, error) {
	// Compare each tree in the forest
	for method, tree := range r.trees {
		otherTree, ok := other.trees[method]
		if !ok {
			return false, fmt.Errorf("cannot find method %s in router\n", method)
		}
		// Compare each node in tree recursively
		ok, err := tree.equal(otherTree)
		if !ok {
			return false, err
		}
	}
	return true, nil

}

// Compare two nodes
func (n *node) equal(that *node) (bool, error) {
	if n.path != that.path {
		return false, fmt.Errorf("different node path: %s vs %s\n", n.path, that.path)
	}
	if len(n.children) != len(that.children) {
		return false, fmt.Errorf("length of children not same: %d vs %d\n", len(n.children), len(that.children))
	}

	// Because functions are not comparable, use reflect value to compare
	nHandler := reflect.ValueOf(n.handler)
	thatHandler := reflect.ValueOf(that.handler)
	if nHandler != thatHandler {
		return false, fmt.Errorf("different handlers")
	}

	for path, c := range n.children {
		dst, ok := that.children[path]
		if !ok {
			return false, fmt.Errorf("child node %s not exist", path)
		}
		ok, err := c.equal(dst)
		if !ok {
			return false, err
		}

	}
	return true, nil
}
