package controlplane

import (
	"context"
	"log"

	controlv1 "github.com/Prateet-Github/sentinel/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	controlv1.UnimplementedSentinelControlServer
	store *Store
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

	log.Printf("data plane connected: %s", req.GetNodeId())

	snapshot := s.store.Snapshot()

	return stream.Send(&controlv1.ConfigResponse{
		Snapshot: snapshot,
	})
}
