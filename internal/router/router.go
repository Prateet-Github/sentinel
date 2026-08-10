package router

import "github.com/Prateet-Github/sentinel/internal/core"

type Router interface {
	Match(method, path string, params *Params) (*core.Route, bool)
}
