package controlplane

import (
	"context"

	controlv1 "github.com/Prateet-Github/sentinel/proto"
)

type Server struct {
	controlv1.UnimplementedSentinelControlServer
}

func (s *Server) ListServices(
	ctx context.Context,
	req *controlv1.ListServicesRequest,
) (*controlv1.ListServicesResponse, error) {
	return &controlv1.ListServicesResponse{}, nil
}
