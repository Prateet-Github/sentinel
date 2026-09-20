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
	storage  *Storage
}

func NewStore(storage *Storage) *Store {
	return &Store{
		services: make(map[string]*controlv1.Service),
		routes:   make(map[string]*controlv1.Route),
		storage:  storage,
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

func (s *Store) AddBackend(
	serviceName string,
	backend *controlv1.Backend,
) (*controlv1.Backend, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	service, exists := s.services[serviceName]

	if !exists {
		if err := s.storage.CreateService(serviceName); err != nil {
			return nil, err
		}

		service = &controlv1.Service{
			Name: serviceName,
		}

		s.services[serviceName] = service
	}

	for _, existing := range service.Backends {
		if existing.Name == backend.Name {
			return nil, fmt.Errorf(
				"backend %q already exists",
				backend.Name,
			)
		}
	}

	if err := s.storage.CreateBackend(serviceName, backend); err != nil {
		return nil, err
	}

	service.Backends = append(service.Backends, backend)

	return backend, nil
}
