package lb

import "sync/atomic"

type SelectionStrategy interface {
	Select(pool *BackendPool) int
}

type RoundRobinStrategy struct {
	next atomic.Uint64
}

func (r *RoundRobinStrategy) Select(pool *BackendPool) int {
	if len(pool.backends) == 0 {
		return -1
	}

	index := r.next.Add(1) - 1

	for i := uint64(0); i < uint64(len(pool.backends)); i++ {
		idx := (index + i) % uint64(len(pool.backends))

		if !pool.Available(int(idx)) {
			continue
		}

		return int(idx)
	}

	return -1
}
