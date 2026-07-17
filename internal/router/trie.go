package router

import (
	"strings"

	"github.com/Prateet-Github/sentinel/internal/core"
)

type TrieNode struct {
	prefix   string
	children []*TrieNode
	route    *core.Route
}

type TrieRouter struct {
	root *TrieNode
}

func NewTrieRouter(cfg *core.Config) *TrieRouter {
	r := &TrieRouter{
		root: &TrieNode{},
	}

	for _, route := range cfg.Routes {
		r.insert(route)
	}

	return r
}

func (r *TrieRouter) Match(method, path string) (*core.Route, bool) {
	node := r.root

	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}

	if len(path) > 0 && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}

	if path == "" {
		if node.route != nil && node.route.Method == method {
			return node.route, true
		}
		return nil, false
	}

	start := 0
	for i := 0; i < len(path); i++ {
		if path[i] == '/' {
			if i > start {
				segment := path[start:i]
				node = findChild(node, segment)
				if node == nil {
					return nil, false
				}
			}
			start = i + 1
		}
	}

	if start < len(path) {
		segment := path[start:]
		node = findChild(node, segment)
		if node == nil {
			return nil, false
		}
	}

	if node.route == nil || node.route.Method != method {
		return nil, false
	}

	return node.route, true
}

func (r *TrieRouter) insert(route core.Route) {
	node := r.root

	segments := strings.Split(strings.Trim(route.Path, "/"), "/")

	for _, segment := range segments {
		child := findChild(node, segment)

		if child == nil {
			child = &TrieNode{
				prefix: segment,
			}
			node.children = append(node.children, child)
		}

		node = child
	}

	node.route = &route
}

func findChild(node *TrieNode, prefix string) *TrieNode {
	for _, child := range node.children {
		if child.prefix == prefix {
			return child
		}
	}

	return nil
}
