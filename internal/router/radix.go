package router

import (
	"strings"

	"github.com/Prateet-Github/sentinel/internal/core"
)

type RadixNode struct {
	prefix   string
	children []*RadixNode
	route    *core.Route
}

type RadixRouter struct {
	root *RadixNode
}

func NewRadixRouter(cfg *core.Config) *RadixRouter {
	r := &RadixRouter{
		root: &RadixNode{},
	}
	for _, route := range cfg.Routes {
		r.insert(route)
	}
	return r
}

func (r *RadixRouter) Match(method, path string) (*core.Route, bool) {
	// normalize leading slash
	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}

	// normalise trailing slash
	if len(path) > 0 && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}

	node := r.root

	for len(path) > 0 {

		var matched *RadixNode

		for _, child := range node.children {
			common := longestCommonPrefix(child.prefix, path)

			// consume the entire child prefix
			if common == len(child.prefix) {
				matched = child
				break
			}
		}

		if matched == nil {
			return nil, false
		}

		// consume the compressed prefix
		path = path[len(matched.prefix):]
		node = matched
	}

	// must end exactly on a route node
	if node.route == nil {
		return nil, false
	}

	// HTTP method must match.
	if node.route.Method != method {
		return nil, false
	}

	return node.route, true
}

func (r *RadixRouter) insert(route core.Route) {
	path := strings.Trim(route.Path, "/")

	if path == "" {
		r.root.route = &route
		return
	}

	r.insertNode(r.root, path, &route)
}

func (r *RadixRouter) insertNode(
	node *RadixNode,
	path string,
	route *core.Route,
) {
	if path == "" {
		node.route = route
		return
	}

	for _, child := range node.children {
		common := longestCommonPrefix(child.prefix, path)

		// Case 1: No common prefix
		if common == 0 {
			continue
		}

		// Case 2: Partial overlap -> split node
		if common < len(child.prefix) {

			oldChild := &RadixNode{
				prefix:   child.prefix[common:],
				children: child.children,
				route:    child.route,
			}

			child.prefix = child.prefix[:common]
			child.children = []*RadixNode{oldChild}
			child.route = nil

			// New route ends at the shared prefix
			if common == len(path) {
				child.route = route
				return
			}

			// Remaining part of the new route becomes a sibling
			newChild := &RadixNode{
				prefix: path[common:],
				route:  route,
			}

			child.children = append(child.children, newChild)
			return
		}

		// Case 3: Exact match
		if common == len(child.prefix) && common == len(path) {
			child.route = route
			return
		}

		// Case 4: Child prefix is a prefix of the path
		remaining := path[common:]

		if len(remaining) > 0 && remaining[0] == '/' {
			remaining = remaining[1:]
		}

		r.insertNode(child, remaining, route)
		return
	}

	// No matching child.
	node.children = append(node.children, &RadixNode{
		prefix: path,
		route:  route,
	})
}

func longestCommonPrefix(a, b string) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}

	i := 0
	for i < n && a[i] == b[i] {
		i++
	}
	return i

}
