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
		{
			name: "parameter route",
			routes: []core.Route{
				{Method: "GET", Path: "/users/:id"},
			},
		},
		{
			name: "static and parameter route",
			routes: []core.Route{
				{Method: "GET", Path: "/users/profile"},
				{Method: "GET", Path: "/users/:id"},
			},
		},
		{
			name: "nested parameter route",
			routes: []core.Route{
				{Method: "GET", Path: "/users/:id/profile"},
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

	fmt.Printf(
		"%s%s\n",
		strings.Repeat("  ", depth),
		node.prefix,
	)

	for _, child := range node.children {
		dump(child, depth+1)
	}

	if node.paramChild != nil {
		fmt.Printf(
			"%s:paramChild\n",
			strings.Repeat("  ", depth+1),
		)

		dump(node.paramChild, depth+2)
	}
}

func TestRadixMatch(t *testing.T) {
	cfg := &core.Config{
		Routes: []core.Route{
			{Method: "GET", Path: "/users"},
			{Method: "GET", Path: "/users/profile"},
			{Method: "GET", Path: "/users/profile/avatar"},
			{Method: "GET", Path: "/users/settings"},

			// Parameterized routes.
			{Method: "GET", Path: "/users/:id"},
			{Method: "GET", Path: "/users/:id/profile"},

			{Method: "GET", Path: "/orders"},
		},
	}

	r := NewRadixRouter(cfg)

	tests := []struct {
		name   string
		method string
		path   string
		want   bool
	}{
		{
			name:   "users",
			method: "GET",
			path:   "/users",
			want:   true,
		},
		{
			name:   "profile",
			method: "GET",
			path:   "/users/profile",
			want:   true,
		},
		{
			name:   "settings",
			method: "GET",
			path:   "/users/settings",
			want:   true,
		},
		{
			name:   "orders",
			method: "GET",
			path:   "/orders",
			want:   true,
		},
		{
			name:   "unknown route",
			method: "GET",
			path:   "/unknown",
			want:   false,
		},
		{
			name:   "wrong method",
			method: "POST",
			path:   "/users",
			want:   false,
		},
		{
			name:   "nested compressed path",
			method: "GET",
			path:   "/users/profile/avatar",
			want:   true,
		},
		{
			name:   "nested parameter route with string",
			method: "GET",
			path:   "/users/pro/profile",
			want:   true,
		},
		{
			name:   "extra path",
			method: "GET",
			path:   "/users/profile/extra",
			want:   false,
		},
		{
			name:   "trailing slash",
			method: "GET",
			path:   "/users/profile/",
			want:   true,
		},

		// Parameterized routes.
		{
			name:   "parameter route",
			method: "GET",
			path:   "/users/42",
			want:   true,
		},
		{
			name:   "parameter route string",
			method: "GET",
			path:   "/users/abc",
			want:   true,
		},
		{
			name:   "nested parameter route",
			method: "GET",
			path:   "/users/42/profile",
			want:   true,
		},
		{
			name:   "nested parameter route string",
			method: "GET",
			path:   "/users/abc/profile",
			want:   true,
		},

		// static route must take priority over :id
		{
			name:   "static beats parameter",
			method: "GET",
			path:   "/users/profile",
			want:   true,
		},

		// /users is a valid static route
		{
			name:   "parameter route missing segment",
			method: "GET",
			path:   "/users",
			want:   true,
		},

		{
			name:   "parameter extra path",
			method: "GET",
			path:   "/users/42/unknown",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ok := r.Match(tt.method, tt.path)

			if ok != tt.want {
				t.Fatalf(
					"Match(%q, %q) = %v, want %v",
					tt.method,
					tt.path,
					ok,
					tt.want,
				)
			}
		})
	}
}
