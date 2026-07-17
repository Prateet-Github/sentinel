package router

import "github.com/Prateet-Github/sentinel/internal/core"

type Node struct {
	prefix   string
	children []*Node
	route    *core.Route
}

type RadixRouter struct {
	root *Node
}

func NewRadixRouter(cfg *core.Config) *RadixRouter {
	r := &RadixRouter{
		root: &Node{},
	}

	// Todo: implement radix tree routing logic here

	return r
}

func (r *RadixRouter) Match(method, path string) (*core.Route, bool) {
	panic("will implement later")
}
