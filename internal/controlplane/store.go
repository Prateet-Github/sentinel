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
		store:       store,
		subscribers: make(map[string]chan *controlv1.ConfigSnapshot),
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

	if exists {
		for _, existing := range service.Backends {
			if existing.Name == backend.Name {
				return nil, fmt.Errorf(
					"backend %q already exists",
					backend.Name,
				)
			}
		}
	}

	// persist everything atomically
	if err := s.storage.AddBackend(serviceName, backend); err != nil {
		return nil, err
	}

	// only update memory after SQLite commit succeeds
	if !exists {
		service = &controlv1.Service{
			Name: serviceName,
		}

		s.services[serviceName] = service
	}

	service.Backends = append(service.Backends, backend)

	return backend, nil
}

func (s *Store) Load() error {
	services, err := s.storage.LoadServices()
	if err != nil {
		return err
	}

	backends, err := s.storage.LoadBackends()
	if err != nil {
		return err
	}

	routes, err := s.storage.LoadRoutes()
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, service := range services {
		service.Backends = backends[service.Name]
		s.services[service.Name] = service
	}

	for _, route := range routes {
		s.routes[routeKey(
			route.GetMethod(),
			route.GetPath(),
		)] = route
	}

	return nil
}

func (s *Store) Snapshot() *controlv1.ConfigSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	services := make([]*controlv1.Service, 0, len(s.services))

	for _, service := range s.services {
		services = append(services, service)
	}

	routes := make([]*controlv1.Route, 0, len(s.routes))

	for _, route := range s.routes {
		routes = append(routes, route)
	}

	return &controlv1.ConfigSnapshot{
		Services: services,
		Routes:   routes,
	}
}

func (s *Store) ListRoutes() []*controlv1.Route {
	s.mu.RLock()
	defer s.mu.RUnlock()

	routes := make(
		[]*controlv1.Route,
		0,
		len(s.routes),
	)

	for _, route := range s.routes {
		routes = append(routes, route)
	}

	return routes
}

func (s *Store) AddRoute(
	route *controlv1.Route,
) (*controlv1.Route, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := routeKey(
		route.GetMethod(),
		route.GetPath(),
	)

	if _, exists := s.routes[key]; exists {
		return nil, fmt.Errorf(
			"route %s %s already exists",
			route.GetMethod(),
			route.GetPath(),
		)
	}

	// Verify that the referenced service exists.
	if _, exists := s.services[route.GetServiceName()]; !exists {
		return nil, fmt.Errorf(
			"service %q not found",
			route.GetServiceName(),
		)
	}

	// Persist first.
	if err := s.storage.AddRoute(route); err != nil {
		return nil, err
	}

	// Update memory only after persistence succeeds.
	s.routes[key] = route

	return route, nil
}

func (s *Store) RemoveRoute(
	method string,
	path string,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := routeKey(method, path)

	if _, exists := s.routes[key]; !exists {
		return fmt.Errorf(
			"route %s %s not found",
			method,
			path,
		)
	}

	if err := s.storage.RemoveRoute(
		method,
		path,
	); err != nil {
		return err
	}

	delete(s.routes, key)

	return nil
}

func routeKey(
	method string,
	path string,
) string {
	return method + " " + path
}
