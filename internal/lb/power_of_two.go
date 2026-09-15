package lb

import "math/rand"

type PowerOfTwoStrategy struct{}

func (p *PowerOfTwoStrategy) Select(pool *BackendPool) int {
	if len(pool.backends) == 0 {
		return -1
	}

	n := len(pool.backends)

	first := rand.Intn(n)
	second := rand.Intn(n)

	for first == second && n > 1 {
		second = rand.Intn(n)
	}

	if !pool.Available(first) {
		return second
	}

	if !pool.Available(second) {
		return first
	}

	if pool.connections[first].Load() <= pool.connections[second].Load() {
		return first
	}

	return second
}
