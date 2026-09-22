package controlplane

import (
	"context"
	"log"
	"sync"

	controlv1 "github.com/Prateet-Github/sentinel/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	controlv1.UnimplementedSentinelControlServer
	store *Store

	mu          sync.RWMutex
	subscribers map[string]chan *controlv1.ConfigSnapshot
}

func (s *Server) ListServices(
	ctx context.Context,
	req *controlv1.ListServicesRequest,
) (*controlv1.ListServicesResponse, error) {
	return &controlv1.ListServicesResponse{
		Services: s.store.ListServices(),
	}, nil
}

func (s *Server) GetService(
	ctx context.Context,
	req *controlv1.GetServiceRequest,
) (*controlv1.GetServiceResponse, error) {
	service, err := s.store.GetService(req.GetName())
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &controlv1.GetServiceResponse{
		Service: service,
	}, nil
}

func (s *Server) AddBackend(
	ctx context.Context,
	req *controlv1.AddBackendRequest,
) (*controlv1.AddBackendResponse, error) {
	backend, err := s.store.AddBackend(
		req.GetServiceName(),
		req.GetBackend(),
	)
	if err != nil {
		return nil, status.Error(codes.AlreadyExists, err.Error())
	}

	s.broadcast(s.store.Snapshot())

	return &controlv1.AddBackendResponse{
		Backend: backend,
	}, nil
}

func (s *Server) StreamConfig(
	stream controlv1.SentinelControl_StreamConfigServer,
) error {
	req, err := stream.Recv()
	if err != nil {
		return err
	}

	nodeID := req.GetNodeId()

	log.Printf("data plane connected: %s", nodeID)

	ch := s.subscribe(nodeID)
	defer s.unsubscribe(nodeID)

	// send current configuration immediately
	if err := stream.Send(&controlv1.ConfigResponse{
		Snapshot: s.store.Snapshot(),
	}); err != nil {
		return err
	}

	// wait for configuration updates
	for {
		select {
		case snapshot := <-ch:
			if err := stream.Send(&controlv1.ConfigResponse{
				Snapshot: snapshot,
			}); err != nil {
				return err
			}

		case <-stream.Context().Done():
			log.Printf("data plane disconnected: %s", nodeID)
			return stream.Context().Err()
		}
	}
}

func (s *Server) subscribe(nodeID string) chan *controlv1.ConfigSnapshot {
	s.mu.Lock()
	defer s.mu.Unlock()

	ch := make(chan *controlv1.ConfigSnapshot, 1)
	s.subscribers[nodeID] = ch

	return ch
}

func (s *Server) unsubscribe(nodeID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if ch, exists := s.subscribers[nodeID]; exists {
		close(ch)
		delete(s.subscribers, nodeID)
	}
}

func (s *Server) broadcast(snapshot *controlv1.ConfigSnapshot) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, ch := range s.subscribers {
		select {
		case ch <- snapshot:
		default:
			// Don't block the Control Plane if a DP hasn't consumed the previous update yet
		}
	}
}
