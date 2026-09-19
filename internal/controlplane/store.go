package controlplane

import (
	"fmt"
	"sync"

	controlv1 "github.com/Prateet-Github/sentinel/proto"
)

type Store struct {
	mu       sync.RWMutex // grpc server can handle multiple requests concurrently gotta protect the store with a mutex
	services map[string]*controlv1.Service
	routes   map[string]*controlv1.Route
}

func NewStore() *Store {
	return &Store{
		services: make(map[string]*controlv1.Service),
		routes:   make(map[string]*controlv1.Route),
	}
}

func NewServer(store *Store) *Server {
	return &Server{
		store: store,
	}
}

func (s *Store) ListServices() []*controlv1.Service {
	s.mu.RLock()
	defer s.mu.RUnlock()

	services := make([]*controlv1.Service, 0, len(s.services))

	for _, service := range s.services {
		services = append(services, service)
	}

	return services
}

func (s *Store) GetService(name string) (*controlv1.Service, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	service, exists := s.services[name]
	if !exists {
		return nil, fmt.Errorf("service %q not found", name)
	}

	return service, nil
}
