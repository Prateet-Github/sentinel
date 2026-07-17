package router

import "github.com/Prateet-Github/sentinel/internal/core"

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
	panic("not implemented")
}

// func longestCommonPrefix(a, b string) int {
// 	n := len(a)
// 	if len(b) < n {
// 		n = len(b)
// 	}

// 	i := 0
// 	for i < n && a[i] == b[i] {
// 		i++
// 	}

// 	return i
// }
