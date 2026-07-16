package router

import "github.com/Prateet-Github/sentinel/internal/core"

type Router interface {
	Match(method, path string) (*core.Route, bool)
}
