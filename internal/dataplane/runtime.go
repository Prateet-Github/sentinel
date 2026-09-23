package dataplane

import (
	"github.com/Prateet-Github/sentinel/internal/core"
	"github.com/Prateet-Github/sentinel/internal/lb"
	"github.com/Prateet-Github/sentinel/internal/router"
)

type RuntimeState struct {
	Config       *core.Config
	Router       *router.RadixRouter
	LoadBalancer *lb.LoadBalancer
}
