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

func TestRadixMatch(t *testing.T) {
	cfg := &core.Config{
		Routes: []core.Route{
			{Method: "GET", Path: "/users"},
			{Method: "GET", Path: "/users/profile"},
			{Method: "GET", Path: "/users/profile/avatar"},
			{Method: "GET", Path: "/users/settings"},
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
			name:   "partial path",
			method: "GET",
			path:   "/users/pro",
			want:   false,
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
