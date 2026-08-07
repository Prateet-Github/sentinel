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

	segments := strings.Split(path, "/")
	r.insertNode(r.root, segments, &route)
}

func (r *RadixRouter) insertNode(
	node *RadixNode,
	segments []string,
	route *core.Route,
) {
	if len(segments) == 0 {
		node.route = route
		return
	}

	path := strings.Join(segments, "/")

	for _, child := range node.children {
		common := longestCommonPrefix(child.prefix, path)

		if common > 0 {
			// TODO: splitting & descending logic for partial matches
			return
		}
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
