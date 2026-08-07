package router

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Prateet-Github/sentinel/internal/core"
)

func TestRadixInsert(t *testing.T) {
	tests := []struct {
		name   string
		routes []core.Route
	}{
		{
			name: "single route",
			routes: []core.Route{
				{Method: "GET", Path: "/users"},
			},
		},
		{
			name: "shared prefix",
			routes: []core.Route{
				{Method: "GET", Path: "/users/profile"},
				{Method: "GET", Path: "/users/settings"},
			},
		},
		{
			name: "different roots",
			routes: []core.Route{
				{Method: "GET", Path: "/users"},
				{Method: "GET", Path: "/orders"},
			},
		},
		{
			name: "nested path",
			routes: []core.Route{
				{Method: "GET", Path: "/users"},
				{Method: "GET", Path: "/users/profile/avatar"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			cfg := &core.Config{
				Routes: tt.routes,
			}

			r := NewRadixRouter(cfg)

			fmt.Println("\n===", tt.name, "===")
			dump(r.root, 0)

			if r.root == nil {
				t.Fatal("root is nil")
			}

			if len(r.root.children) == 0 {
				t.Fatal("expected children")
			}
		})
	}
}

func dump(node *RadixNode, depth int) {
	if node == nil {
		return
	}

	fmt.Printf("%s%s\n",
		strings.Repeat("  ", depth),
		node.prefix,
	)

	for _, child := range node.children {
		dump(child, depth+1)
	}
}
