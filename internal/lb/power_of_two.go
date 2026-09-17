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

	firstAvailable := pool.Available(first)
	secondAvailable := pool.Available(second)

	if !firstAvailable && !secondAvailable {
		return -1
	}

	if !firstAvailable {
		return second
	}

	if !secondAvailable {
		return first
	}

	if pool.connections[first].Load() <= pool.connections[second].Load() {
		return first
	}

	return second
}
