package lb

import "math"

type LeastConnectionStrategy struct{}

func (l *LeastConnectionStrategy) Select(pool *BackendPool) int {

	if len(pool.backends) == 0 {
		return -1
	}

	selected := -1
	minConnections := int64(math.MaxInt64)

	for i := range pool.backends {

		if !pool.Available(i) {
			continue
		}
		connections := pool.connections[i].Load()
		if connections < minConnections {
			minConnections = connections
			selected = i
		}
	}

	return selected

}
