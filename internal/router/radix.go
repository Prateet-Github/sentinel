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

func (r *RadixRouter) Match(method, path string, params *Params) (*core.Route, bool) {
	path = strings.Trim(path, "/")

	node := r.root

	for len(path) > 0 {
		var matched *RadixNode

		// 1. Static routes have priority.
		for _, child := range node.children {
			common := longestCommonPrefix(child.prefix, path)

			if common != len(child.prefix) {
				continue
			}

			// The match must end on a segment boundary.
			if len(path) > len(child.prefix) &&
				child.prefix[len(child.prefix)-1] != '/' &&
				path[len(child.prefix)] != '/' {
				continue
			}

			matched = child
			break
		}

		// 2. Fall back to parameter route.
		if matched == nil && node.paramChild != nil {
			matched = node.paramChild
		}

		if matched == nil {
			return nil, false
		}

		// 3. Parameter route.
		if matched == node.paramChild {
			slash := strings.IndexByte(path, '/')

			var value string

			if slash == -1 {
				value = path
				path = ""
			} else {
				value = path[:slash]
				path = path[slash+1:]
			}

			name := strings.TrimPrefix(matched.prefix, ":")

			// Append directly to caller's stack-backed slice (ZERO ALLOCATIONS)
			if params != nil {
				*params = append(*params, Param{Key: name, Value: value})
			}

			node = matched
			continue
		}

		// 4. Static compressed route.
		path = path[len(matched.prefix):]

		if len(path) > 0 {
			if path[0] == '/' {
				path = path[1:]
			}
		}

		node = matched
	}

	// 5. Final route validation.
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

	// No matching static child.
	//
	// Only store the first path segment in the radix node.
	// The remaining path is inserted recursively so that
	// parameter segments can be represented by paramChild.
	// segments := strings.SplitN(path, "/", 2)

	child := &RadixNode{
		prefix: segments[0],
	}

	node.children = append(node.children, child)

	if len(segments) == 1 {
		child.route = route
		return
	}

	r.insertNode(child, segments[1], route)
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
