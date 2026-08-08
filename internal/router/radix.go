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
	path = strings.Trim(path, "/")
	node := r.root

	for len(path) > 0 {

		var matched *RadixNode

		for _, child := range node.children {

			common := longestCommonPrefix(child.prefix, path)

			if common == len(child.prefix) {
				matched = child
				break
			}
		}

		if matched == nil {

			return nil, false
		}

		path = path[len(matched.prefix):]
		node = matched

		if len(path) > 0 {
			if path[0] != '/' {
				return nil, false
			}
			path = path[1:]
		}
	}

	if node.route == nil || node.route.Method != method {
		return nil, false
	}

	return node.route, true
}

func (r *RadixRouter) insert(route core.Route) {
	path := strings.Trim(route.Path, "/")

	if path == "" {
		routeCopy := route
		r.root.route = &routeCopy
		return
	}

	routeCopy := route
	r.insertNode(r.root, path, &routeCopy)
}

func (r *RadixRouter) insertNode(node *RadixNode, path string, route *core.Route) {
	if path == "" {
		node.route = route
		return
	}

	for _, child := range node.children {
		common := longestCommonPrefix(child.prefix, path)

		if common == 0 {
			continue
		}

		// Case 1: Partial overlap: Split existing node
		if common < len(child.prefix) {
			oldChild := &RadixNode{
				prefix:   child.prefix[common:],
				children: child.children,
				route:    child.route,
			}

			// Clean leading slash on split child prefix
			if len(oldChild.prefix) > 0 && oldChild.prefix[0] == '/' {
				oldChild.prefix = oldChild.prefix[1:]
			}

			child.prefix = child.prefix[:common]
			if len(child.prefix) > 0 && child.prefix[len(child.prefix)-1] == '/' {
				child.prefix = child.prefix[:len(child.prefix)-1]
			}

			child.children = []*RadixNode{oldChild}
			child.route = nil

			// Route ends exactly at split boundary
			if common == len(path) {
				child.route = route
				return
			}

			// Attach leftover segment as sibling
			remaining := path[common:]
			if len(remaining) > 0 && remaining[0] == '/' {
				remaining = remaining[1:]
			}

			newChild := &RadixNode{
				prefix: remaining,
				route:  route,
			}
			child.children = append(child.children, newChild)
			return
		}

		// Case 2: Exact match on child prefix
		if common == len(child.prefix) && common == len(path) {
			child.route = route
			return
		}

		// Case 3: Descend down tree
		remaining := path[common:]
		if len(remaining) > 0 && remaining[0] == '/' {
			remaining = remaining[1:]
		}

		r.insertNode(child, remaining, route)
		return
	}

	// Case 4: No overlap: Append new child leaf
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
