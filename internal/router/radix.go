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
	panic("not implemented")
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

		// Case 2: Exact match
		if common == len(child.prefix) &&
			common == len(path) {

			child.route = route
			return
		}

		// Case 3: Child prefix is a prefix of the path

		if common == len(child.prefix) {
			remaining := path[common:]

			if len(remaining) > 0 && remaining[0] == '/' {
				remaining = remaining[1:]
			}

			r.insertNode(child, remaining, route)
			return
		}

		// Case 4: Partial overlap

	}

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
