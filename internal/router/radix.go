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
	r.insertNode(r.root, path, &route)
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

func (r *RadixRouter) insertNode(
	node *RadixNode,
	path string,
	route *core.Route,
) {
	// TODO: implement radix tree insertion logic here

	for _, child := range node.children {
		if longestCommonPrefix(child.prefix, path) > 0 {
			// will handle this next
			return
		}
	}

	node.children = append(node.children, &RadixNode{
		prefix: path,
		route:  route,
	})
}
