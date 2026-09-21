package dataplane

import (
	"context"
	"fmt"

	controlv1 "github.com/Prateet-Github/sentinel/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ControlClient struct {
	client controlv1.SentinelControlClient
	conn   *grpc.ClientConn
}

func NewControlClient(
	ctx context.Context,
	address string,
) (*ControlClient, error) {
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	return &ControlClient{
		client: controlv1.NewSentinelControlClient(conn),
		conn:   conn,
	}, nil
}

func (c *ControlClient) StreamConfig(
	ctx context.Context,
	nodeID string,
) (*controlv1.ConfigSnapshot, error) {
	stream, err := c.client.StreamConfig(ctx)
	if err != nil {
		return nil, err
	}

	if err := stream.Send(&controlv1.ConfigRequest{
		NodeId: nodeID,
	}); err != nil {
		return nil, err
	}

	response, err := stream.Recv()
	if err != nil {
		return nil, err
	}

	if response.GetSnapshot() == nil {
		return nil, fmt.Errorf("received empty config snapshot")
	}

	return response.GetSnapshot(), nil
}

func (c *ControlClient) Close() error {
	return c.conn.Close()
}
