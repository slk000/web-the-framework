package web

import (
    "fmt"
    "regexp"
    "strings"
)

// :a:b((\d)(.*))
// ::b:()((()
// :123()
// :()
// :(123)
// :(()
// :())
// :3f
var regexpRoutePattern, _ = regexp.Compile("((?::.*?)+)\\((.*)\\)") // ((?::.*?)+)\((.*)\) capture 1.keys 2.regexp
// node of a router tree
type node struct {
    path          string           // sub-path of this node
    children      map[string]*node // normal children nodes
    regexpChild   *node            // regex expression child node
    regexp        *regexp.Regexp
    wildcardChild *node // wildcardChild child node
    paramChild    *node // param child node
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
        // Priority: static > regexp > param > wildcard
        if root.children[subPath] != nil {
            root = root.children[subPath]
        } else if root.regexp != nil && root.regexp.MatchString(subPath) {
            keysStr := regexpRoutePattern.FindStringSubmatch(root.regexpChild.path)[1] // keys are stored in path
            var keys []string
            if keysStr != ":" {
                keys = strings.Split(keysStr[1:], ":") // strings.Split("a:b:c:d",":") ---> ["a","b,"c,"d"].
            }
            values := root.regexp.FindStringSubmatch(subPath)
            // if matched, values[1...] is the captured parts, [0] is the entire string (subPath)
            // so omit [0]
            if len(values)-1 != len(keys) {
                //panic("regexp len(keys) != values(keys)")
                fmt.Println("regexp len(keys) != values(keys)")
                return nil, nil
            }

            if params == nil {
                params = make(map[string]string)
            }
            for idx := 0; idx < len(keys); idx++ {
                params[keys[idx]] = values[idx+1]
            }
            root = root.regexpChild
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
    //    return nil
    //}

    // is a param child or a regex child
    if subPath[0] == ':' {
        matches := regexpRoutePattern.FindStringSubmatch(subPath)
        // If subPath is a regex route, matches will get 3 parts:
        //        1. matches[0] == subPath
        //        2. matches[1]: keys, ":key1:key2:key3..."
        //        3. matches[2]: the user provided regex to match the route and capture values
        // If not, subPath is a param route
        if len(matches) == 3 {
            // a regex child
            if n.regexpChild == nil {
                n.regexp = regexp.MustCompile(matches[2])
                n.regexpChild = &node{path: subPath}
            }
            return n.regexpChild
        }

        // a param child
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
