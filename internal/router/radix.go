package router

import (
	"strings"

	"github.com/Prateet-Github/sentinel/internal/core"
)

type RadixNode struct {
	prefix     string
	children   []*RadixNode
	paramChild *RadixNode // for parameterized routes
	route      *core.Route
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

		// Static routes always get priority.
		for _, child := range node.children {
			common := longestCommonPrefix(child.prefix, path)

			if common == len(child.prefix) {
				matched = child
				break
			}
		}

		// If no static route matches, try parameterized route
		if matched == nil && node.paramChild != nil {
			matched = node.paramChild
		}

		if matched == nil {
			return nil, false
		}

		// Parameterized route consumes exactly one path segment
		if matched == node.paramChild {
			slash := strings.IndexByte(path, '/')

			if slash == -1 {
				path = ""
			} else {
				path = path[slash+1:]
			}

			node = matched
			continue
		}

		// Static compressed prefix
		path = path[len(matched.prefix):]

		// If more path remains, the next character must be
		// a segment separator
		if len(path) > 0 {
			if path[0] != '/' {
				return nil, false
			}

			path = path[1:]
		}

		node = matched
	}

	if node.route == nil || node.route.Method != method {
		return nil, false
	}

	return node.route, true
}

func (r *RadixRouter) insert(route core.Route) {
	path := strings.Trim(route.Path, "/")

	routeCopy := route

	if path == "" {
		r.root.route = &routeCopy
		return
	}

	r.insertNode(r.root, path, &routeCopy)
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

	// Parameterized route
	segments := strings.SplitN(path, "/", 2)

	if strings.HasPrefix(segments[0], ":") {
		if node.paramChild == nil {
			node.paramChild = &RadixNode{
				prefix: segments[0],
			}
		}

		if len(segments) == 1 {
			node.paramChild.route = route
			return
		}

		r.insertNode(node.paramChild, segments[1], route)
		return
	}

	// Static route
	for _, child := range node.children {
		common := longestCommonPrefix(child.prefix, path)

		if common == 0 {
			continue
		}

		// Partial overlap -> split existing node
		if common < len(child.prefix) {
			oldChild := &RadixNode{
				prefix:     child.prefix[common:],
				children:   child.children,
				paramChild: child.paramChild,
				route:      child.route,
			}

			child.prefix = child.prefix[:common]
			child.children = []*RadixNode{oldChild}
			child.paramChild = nil
			child.route = nil

			// New route ends at the split point
			if common == len(path) {
				child.route = route
				return
			}

			// Remaining part becomes a new sibling
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

		// Exact match
		if common == len(child.prefix) && common == len(path) {
			child.route = route
			return
		}

		// Existing child is a prefix of the new path
		remaining := path[common:]

		if len(remaining) > 0 && remaining[0] == '/' {
			remaining = remaining[1:]
		}

		r.insertNode(child, remaining, route)
		return
	}

	// No matching static child
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
