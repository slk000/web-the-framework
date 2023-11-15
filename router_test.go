package web

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"reflect"
	"regexp"
	"testing"
)

/*
*****************
    AddRoute
*****************
*/
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
	           /          \
	   *"user"            "order"
	      |               |
	   *"home"            *"detail"
	*/

	// The expected router structure
	home := &node{
		path:    "home",
		handler: mockHandler,
	}
	detail := &node{
		path:    "detail",
		handler: mockHandler,
	}
	order := &node{
		path:     "order",
		children: map[string]*node{"detail": detail},
	}
	user := &node{
		path:     "user",
		handler:  mockHandler,
		children: map[string]*node{"home": home},
	}
	root := &node{
		path:     "/",
		handler:  mockHandler,
		children: map[string]*node{"user": user, "order": order},
	}

	wantRouter := &router{
		trees: map[string]*node{
			http.MethodGet: root,
		},
	}

	// Test
	ok, err := wantRouter.equal(r)
	assert.True(t, ok, err)
	ok, err = r.equal(wantRouter)
	assert.True(t, ok, err)
}

// Test add incorrect static paths
func TestRouter_AddRouteStaticIncorrect(t *testing.T) {
	var mockHandler = func(ctx *Context) {}
	r := newRouter()

	// Empty path is not allowed
	assert.Panicsf(t, func() {
		r.AddRoute(http.MethodGet, "", mockHandler)
	}, "Empty path is not allowed")

	// Path should not end with '/'
	assert.Panicsf(t, func() {
		r.AddRoute(http.MethodGet, "/a/", mockHandler)
	}, "Path should not end with '/'")

	// Path should start with '/'
	assert.Panicsf(t, func() {
		r.AddRoute(http.MethodGet, "d/d", mockHandler)
	}, "Path should start with '/'")

	// Path should not contain continuous '/'
	assert.Panicsf(t, func() {
		r.AddRoute(http.MethodGet, "/d//d", mockHandler)
	}, "Path should not contain continuous '/'")

	// Duplicate node: root
	r.AddRoute(http.MethodGet, "/", mockHandler)
	assert.Panicsf(t, func() {
		r.AddRoute(http.MethodGet, "/", mockHandler)
	}, "Duplicate node: root")

	// Duplicate node
	r.AddRoute(http.MethodGet, "/a", mockHandler)
	assert.Panicsf(t, func() {
		r.AddRoute(http.MethodGet, "/a", mockHandler)
	}, "Duplicate node")
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
		{
			method: http.MethodGet,
			path:   "/",
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
	                   /          \
	           *"user"             *"*"
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
	wildcardChildOfUser := &node{
		path:     "*",
		children: map[string]*node{"home": home2},
	}
	user := &node{
		path:          "user",
		handler:       mockHandler,
		children:      map[string]*node{"nobody": nobody},
		wildcardChild: wildcardChildOfUser,
	}
	wildcardChildOfRoot := &node{
		path:    "*",
		handler: mockHandler,
	}
	wildcardChildOfRoot.wildcardChild = wildcardChildOfRoot // path "/*" has a trailing wildcard
	root := &node{
		path:          "/",
		children:      map[string]*node{"user": user},
		wildcardChild: wildcardChildOfRoot,
		handler:       mockHandler,
	}

	wantRouter := &router{
		trees: map[string]*node{
			http.MethodGet: root,
		},
	}

	// Test
	ok, err := wantRouter.equal(r)
	assert.True(t, ok, err)
	ok, err = r.equal(wantRouter)
	assert.True(t, ok, err)
}

func TestRouter_AddRouteWildcardIncorrect(t *testing.T) {
	var mockHandler = func(ctx *Context) {}
	r := newRouter()
	assert.NotPanics(t, func() {
		r.AddRoute(http.MethodGet, "/home/*", mockHandler)
	})
	assert.Panicsf(t, func() {
		r.AddRoute(http.MethodGet, "/home/*", mockHandler)
	}, "Duplicate wildcard node")
}

// Test add correct param paths
func TestRouter_AddRouteParamCorrect(t *testing.T) {
	// The paths
	testRoutes := []struct {
		method string
		path   string
	}{
		{
			method: http.MethodGet,
			path:   "/:msg",
		},
		{
			method: http.MethodGet,
			path:   "/user",
		},
		{
			method: http.MethodGet,
			path:   "/user/:id/home",
		},
		{
			method: http.MethodGet,
			path:   "/user/nobody/home",
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
	                   /          \
	           *"user"             *":msg"
	            /  \
	        "nobody" ":id"
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
	paramChildOfUser := &node{
		path:     ":id",
		children: map[string]*node{"home": home2},
	}
	user := &node{
		path:       "user",
		handler:    mockHandler,
		children:   map[string]*node{"nobody": nobody},
		paramChild: paramChildOfUser,
	}
	paramChildOfRoot := &node{
		path:    ":msg",
		handler: mockHandler,
	}
	root := &node{
		path:       "/",
		children:   map[string]*node{"user": user},
		paramChild: paramChildOfRoot,
	}

	wantRouter := &router{
		trees: map[string]*node{
			http.MethodGet: root,
		},
	}

	// Test
	ok, err := wantRouter.equal(r)
	assert.True(t, ok, err)
	ok, err = r.equal(wantRouter)
	assert.True(t, ok, err)
}

func TestRouter_AddRouteParamIncorrect(t *testing.T) {
	var mockHandler = func(ctx *Context) {}
	r := newRouter()
	assert.NotPanics(t, func() {
		r.AddRoute(http.MethodGet, "/home/:id", mockHandler)
	})
	assert.Panicsf(t, func() {
		r.AddRoute(http.MethodGet, "/home/:id", mockHandler)
	}, "Duplicate param node")
}

func TestRouter_AddRouteRegexpCorrect(t *testing.T) {
	testRoutes := []struct {
		method string
		path   string
	}{
		{
			method: http.MethodGet,
			path:   "/user/:role((.*)_.*)", // "/user/admin_a" -> role=admin
		},
		{
			method: http.MethodGet,
			path:   "/user/:role((.*)_.*)/home",
		},
		{
			method: http.MethodPost,
			path:   "/validFormat/a/:(.*)",
		},
		{
			method: http.MethodPost,
			path:   "/validFormat/b/:()",
		},
		{
			method: http.MethodPost,
			path:   "/testParamChild/:paramChild",
		},
		//{
		//    method: http.MethodGet,
		//    path:   "/user/:role(AdminA)",
		//},
		//{
		//    method: http.MethodGet,
		//    path:   "/user/:role(AdminB)/home",
		//},
		//{
		//    method: http.MethodGet,
		//    path:   "/user/:role(AdminC)",
		//},
		//{
		//    method: http.MethodGet,
		//    path:   "/user/:role(AdminC)/home",
		//},
		//{
		//    method: http.MethodGet,
		//    path:   "/user/:role(UserA)",
		//},
		//{
		//    method: http.MethodGet,
		//    path:   "/user/:role(UserA)/home",
		//},
		{
			method: http.MethodGet,
			path:   "/:id((\\d+))",
		},
	}
	var mockHandler = func(ctx *Context) {}

	// Construct the router
	r := newRouter()
	for _, route := range testRoutes {
		r.AddRoute(route.method, route.path, mockHandler)
	}

	/* GET
	             "/"
	         /         \
	     "user"       *":id(\d*)"
	       |
	   *":role((.*)_.*)"
	       |
	      *home
	*/
	home := &node{
		path:    "home",
		handler: mockHandler,
	}
	role := &node{
		path:     ":role((.*)_.*)",
		children: map[string]*node{"home": home},
	}
	user := &node{
		path:        "user",
		regexpChild: role,
		regexp:      regexp.MustCompile(":(.*)_.*"),
	}
	id := &node{
		path:    ":id(\\d+)",
		handler: mockHandler,
	}
	getRoot := &node{
		path:        "/",
		children:    map[string]*node{"user": user},
		regexpChild: id,
		regexp:      regexp.MustCompile("\\d+"),
	}

	/* POST
	                    "/"
	            /                   \
	      "validFormat"      "testParamChild"
	      /            \            \
	     "a"          "b"            *":paramChild"
	      |            |
	   *":(.*)"      *":()"
	*/
	regexpChildOfA := &node{
		path:    ":(.*)",
		handler: mockHandler,
	}
	regexpChildOfB := &node{
		path:    ":()",
		handler: mockHandler,
	}
	a := &node{
		path:        "a",
		regexp:      regexp.MustCompile(".*"),
		regexpChild: regexpChildOfA,
	}
	b := &node{
		path:        "b",
		regexp:      regexp.MustCompile(""),
		regexpChild: regexpChildOfB,
	}
	validFormat := &node{
		path:     "validFormat",
		children: map[string]*node{"a": a, "b": b},
	}

	paramChild := &node{
		path:    ":paramChild",
		handler: mockHandler,
	}
	testParamChild := &node{
		path:       "testParamChild",
		paramChild: paramChild,
	}
	postRoot := &node{
		path:     "/",
		children: map[string]*node{"validFormat": validFormat, "testParamChild": testParamChild},
	}

	wantRouter := &router{
		trees: map[string]*node{
			http.MethodGet:  getRoot,
			http.MethodPost: postRoot,
		},
	}
	// Test
	ok, err := wantRouter.equal(r)
	assert.True(t, ok, err)
	ok, err = r.equal(wantRouter)
	assert.True(t, ok, err)

}

func TestRouter_AddRouteRegexpIncorrect(t *testing.T) {
	var mockHandler = func(ctx *Context) {}
	r := newRouter()
	assert.NotPanics(t, func() {
		r.AddRoute(http.MethodGet, "/home/:a(.+)", mockHandler)
	})
	assert.Panicsf(t, func() {
		r.AddRoute(http.MethodGet, "/home/:b(.*)", mockHandler)
	}, "Duplicate regexp node")

	// Incorrect regex expression
	assert.Panicsf(t, func() {
		r.AddRoute(http.MethodGet, "/a/:a(\\)", mockHandler)
	}, "Invalid regexp")

}

/*
*****************
    FindRoute
*****************
*/
// Test: find route node by static path
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
			caseName: "find root",
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
			foundNode, _ := r.FindRoute(tc.method, tc.fullPath)
			if tc.wantedNode == nil {
				assert.Nil(t, foundNode)
			} else {
				ok, err := tc.wantedNode.equal(foundNode)
				assert.True(t, ok, err)
			}
		})
	}
}

// Test: find route node by wildcard path
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
			path:   "/*", // no child other than wildcardChild
		},
		// test trailing wildcard
		{
			method: http.MethodPut,
			path:   "/a/*", // test: when route path ends with wildcard, match all things after. eg. /a/(b) /a/(b/c/d)
		},
		{
			method: http.MethodPut,
			path:   "/a/*/b", // more specific route even there is a general "/a/*"
		},
		{
			method: http.MethodPut,
			path:   "/aa/*/bb", // not trailing
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
			caseName: "a node has no child but a wildcardChild",
			method:   http.MethodPost,
			fullPath: "/bruh",
			wantedNode: &node{
				path:    "*",
				handler: mockHandler,
			},
		},
		// next 3 cases: when route path ends with wildcard, match all things after.
		{
			caseName: "trilling wildcard: '/a/*'",
			method:   http.MethodPut,
			fullPath: "/a/b/c/d/ef",
			wantedNode: &node{
				path:    "*",
				handler: mockHandler,
				children: map[string]*node{
					"b": &node{
						path:    "b",
						handler: mockHandler,
					},
				},
			},
		},
		{
			caseName: "more specific route '/a/*/b' even there is a trailing wildcard parent '/a/*'",
			method:   http.MethodPut,
			fullPath: "/a/whatever/b",
			wantedNode: &node{
				path:    "b",
				handler: mockHandler,
			},
		},
		{
			caseName: "if more specific path not match, fallback to '/a/*'",
			method:   http.MethodPut,
			fullPath: "/a/whatever/bb",
			wantedNode: &node{
				path:    "*",
				handler: mockHandler,
				children: map[string]*node{
					"b": &node{
						path:    "b",
						handler: mockHandler,
					},
				},
			},
		},
		{
			caseName:   "more specific route but not match '/a/*/b",
			method:     http.MethodPut,
			fullPath:   "/a/whatever/b/c",
			wantedNode: nil, // do not fallback to "/a/*"
		},
		{
			caseName: "not trilling wildcard1",
			method:   http.MethodPut,
			fullPath: "/aa/*/bb",
			wantedNode: &node{
				path:    "bb",
				handler: mockHandler,
			},
		},
		{
			caseName:   "not trilling wildcard2",
			method:     http.MethodPut,
			fullPath:   "/aa/*/cc",
			wantedNode: nil,
		},
	}

	// run sub-testcases
	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			foundNode, _ := r.FindRoute(tc.method, tc.fullPath)
			if tc.wantedNode == nil {
				assert.Nil(t, foundNode)
			} else {
				ok, err := tc.wantedNode.equal(foundNode)
				assert.True(t, ok, err)
			}
		})
	}
}

func TestRouter_FindRouteRegexp(t *testing.T) {
	// The paths to construct router
	testRoutes := []struct {
		method string
		path   string
	}{
		{
			method: http.MethodGet,
			path:   "/user/:role((.*)_.*)", // "/user/admin_a" -> role=admin
		},
		{
			method: http.MethodGet,
			path:   "/user/:role((.*)_.*)/home",
		},
		{
			method: http.MethodPost,
			path:   "/validFormat/a/:(.*)",
		},
		{
			method: http.MethodPost,
			path:   "/validFormat/b/:()",
		},
		{
			method: http.MethodPost,
			path:   "/testParamChild/:paramChild",
		},
		{
			method: http.MethodPut,
			path:   "/:id((\\d+))",
		},
	}

	var mockHandler = func(ctx *Context) {}

	r := newRouter()
	for _, route := range testRoutes {
		r.AddRoute(route.method, route.path, mockHandler)
	}

	// Expected nodes
	home := &node{
		path:    "home",
		handler: mockHandler,
	}
	role := &node{
		path:     ":role((.*)_.*)",
		handler:  mockHandler,
		children: map[string]*node{"home": home},
	}
	// Cases of path
	testCases := []struct {
		caseName   string
		method     string
		fullPath   string
		wantedNode *node // Expected node
	}{
		{
			caseName:   "test match /user/:role((.*)_.*)",
			method:     http.MethodGet,
			fullPath:   "/user/admin_abc",
			wantedNode: role,
		},
		{
			caseName:   "test not match /user/:role((.*)_.*)",
			method:     http.MethodGet,
			fullPath:   "/user/admin",
			wantedNode: nil,
		},
		{
			caseName: "test match /user/:role((.*)_.*)/home",
			method:   http.MethodGet,
			fullPath: "/user/admin_abc/home",
			wantedNode: &node{
				path:    "home",
				handler: mockHandler,
			},
		},
		{
			caseName:   "test not match /user/:role((.*)_.*)/home",
			method:     http.MethodGet,
			fullPath:   "/user/admin/home",
			wantedNode: nil,
		},
		{
			caseName: "test match /validFormat/a/:(.*)",
			method:   http.MethodPost,
			fullPath: "/validFormat/a/abcdef",
			wantedNode: &node{
				path:    ":(.*)",
				handler: mockHandler,
			},
		},
		{
			caseName: "test match /:id((\\d+))", // /:id((\d*))
			method:   http.MethodPut,
			fullPath: "/1234",
			wantedNode: &node{
				path:    ":id((\\d+))",
				handler: mockHandler,
			},
		},
		{
			caseName:   "test not match /:id((\\d+))", // /:id((\d*))
			method:     http.MethodPut,
			fullPath:   "/notNumber",
			wantedNode: nil,
		},
	}

	// run sub-testcases
	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			foundNode, _ := r.FindRoute(tc.method, tc.fullPath)
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
	if n.wildcardChild != nil {
		var ok bool
		var err error
		// if this node is a trailing wildcard, do not compare recursively
		if n.wildcardChild == n {
			if that.wildcardChild != that {
				// that.wildcardChild should also point to itself
				return false, errors.New("trailing wildcard child not point to itself")
			}
		} else {
			ok, err = n.wildcardChild.equal(that.wildcardChild)
			if !ok {
				return ok, err
			}
		}
	}

	if n.paramChild != nil {
		ok, err := n.paramChild.equal(that.paramChild)
		if !ok {
			return ok, err
		}
	}

	// Because functions are not comparable, use reflect value to compare
	nHandler := reflect.ValueOf(n.handler)
	thatHandler := reflect.ValueOf(that.handler)
	if nHandler != thatHandler {
		return false, fmt.Errorf("handlers not match")
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
