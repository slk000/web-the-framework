package web

import "strings"

// node of a router tree
type node struct {
	path     string
	children map[string]*node
	wildcard *node // wildcard child node
	handler  HandleFunc
}

// router tree (actually router forest)
type router struct {
	trees map[string]*node // methods trees
}

// newRouter Creates a new router
func newRouter() *router {
	return &router{
		trees: make(map[string]*node),
	}
}

// AddRoute adds a route in the router of method
// Path limitation: start with '/', end without '/', no continuous '/'
func (r *router) AddRoute(method string, path string, handleFunc HandleFunc) {
	// Validate path
	if path == "" {
		panic("empty path")
	}
	if path[0] != '/' {
		panic("path not start with /")
	}
	if path != "/" && path[len(path)-1] == '/' {
		panic("path end with /")
	}

	// Get the router tree of method
	root, ok := r.trees[method]
	if !ok {
		// Root node not exist, create. Assume its path is '/'
		root = &node{
			path: "/",
		}
		r.trees[method] = root
	}

	if path == "/" {
		if root.handler != nil {
			panic("Duplicate root node")
		}
		root.handler = handleFunc
		return
	}

	path = path[1:] // Remove leading '/', or "/a/b" will be Split to ["", "a", "b"]
	segs := strings.Split(path, "/")
	for _, seg := range segs {
		if seg == "" {
			panic("Continuous '/' in path")
		}
		children := root.getOrCreateChild(seg)
		root = children
	}
	if root.handler != nil {
		panic("Duplicate node")
	}
	root.handler = handleFunc
}

// FindRoute finds a node of given method and path
func (r *router) FindRoute(method, path string) *node {
	root, ok := r.trees[method]
	if !ok {
		// No such method
		return nil
	}

	// root
	if path == "/" {
		return root
	}

	path = strings.Trim(path, "/") // Remove leading and trailing '/'
	segs := strings.Split(path, "/")
	for _, seg := range segs {
		// find in normal children first
		nextNode := root.children[seg]

		// if not found, then check the wildcard child
		if nextNode == nil {
			nextNode = root.wildcard
		}

		// 404
		if nextNode == nil {
			return nil
		}
		root = nextNode
	}
	return root
}

// getOrCreateChild gets n's child node whose sub-path is seg. If not exist, create.
func (n *node) getOrCreateChild(seg string) *node {
	// init children nodes map
	if n.children == nil {
		n.children = map[string]*node{}
	}

	// is a wildcard child
	if seg == "*" {
		if n.wildcard == nil {
			n.wildcard = &node{path: "*"}
		}
		return n.wildcard
	}

	// is a static child
	res, ok := n.children[seg]
	if !ok {
		res = &node{path: seg}
		n.children[seg] = res
	}
	return res
}
