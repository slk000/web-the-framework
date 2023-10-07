package web

import "strings"

// node of a router tree
type node struct {
	path          string           // sub-path of this node
	children      map[string]*node // normal children nodes
	wildcardChild *node            // wildcardChild child node
	paramChild    *node            // param child node
	handler       HandleFunc
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
		panic("path should start with /")
	}
	if path != "/" && path[len(path)-1] == '/' {
		panic("path should not end with /")
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
	subPaths := strings.Split(path, "/")
	for _, subPath := range subPaths {
		if subPath == "" {
			panic("Continuous '/' in path")
		}
		child := root.getOrCreateChild(subPath)
		root = child
	}
	if root.handler != nil {
		panic("Duplicate node")
	}
	// for trailing wildcard, I make it points to itself, so that can pairs anything left
	//
	if subPaths[len(subPaths)-1] == "*" {
		root.wildcardChild = root
	}
	root.handler = handleFunc
}

// FindRoute finds a node of given method and path
func (r *router) FindRoute(method, path string) (*node, *param) {
	root, ok := r.trees[method]
	if !ok {
		// No such method
		return nil, nil
	}

	// root
	if path == "/" {
		return root, nil
	}

	path = strings.Trim(path, "/") // Remove leading and trailing '/'
	subPaths := strings.Split(path, "/")
	var params param
	for _, subPath := range subPaths {
		// Priority: static > param > wildcard
		if root.children[subPath] != nil {
			root = root.children[subPath]
		} else if root.paramChild != nil {
			if params == nil {
				params = make(map[string]string)
			}
			//TODO check duplication
			params[root.paramChild.path[1:]] = subPath
			root = root.paramChild
		} else if root.wildcardChild != nil {
			root = root.wildcardChild
		} else {
			// 404
			return nil, nil
		}
	}
	return root, &params
}

// getOrCreateChild gets n's child node whose sub-path is subPath. If not exist, create.
func (n *node) getOrCreateChild(subPath string) *node {
	//if len(subPath) == 0 {
	//	return nil
	//}

	// is a param child
	if subPath[0] == ':' {
		if n.paramChild == nil {
			n.paramChild = &node{path: subPath}
		}
		return n.paramChild
	}

	// is a wildcard child
	if subPath == "*" {
		if n.wildcardChild == nil {
			n.wildcardChild = &node{path: "*"}
		}
		return n.wildcardChild
	}

	// is a static child
	// init children nodes map
	if n.children == nil {
		n.children = map[string]*node{}
	}
	res, ok := n.children[subPath]
	if !ok {
		res = &node{path: subPath}
		n.children[subPath] = res
	}
	return res
}
