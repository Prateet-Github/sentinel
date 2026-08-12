package lb

import (
	"sync/atomic"

	"github.com/Prateet-Github/sentinel/internal/core"
)

type BackendPool struct {
	backends []*core.Backend
	next     atomic.Uint64
}

func NewBackendPool(backends []*core.Backend) *BackendPool {
	return &BackendPool{
		backends: backends,
	}
}

func (p *BackendPool) Next() *core.Backend {
	if len(p.backends) == 0 {
		return nil
	}

	index := p.next.Add(1) - 1
	return p.backends[index%uint64(len(p.backends))]
}
