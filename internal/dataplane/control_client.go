package dataplane

import (
	"context"
	"fmt"
	"log"

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
	runtimeConfig *RuntimeConfig,
) error {
	stream, err := c.client.StreamConfig(ctx)
	if err != nil {
		return err
	}

	if err := stream.Send(&controlv1.ConfigRequest{
		NodeId: nodeID,
	}); err != nil {
		return err
	}

	for {
		response, err := stream.Recv()
		if err != nil {
			return err
		}

		snapshot := response.GetSnapshot()
		if snapshot == nil {
			return fmt.Errorf("received empty config snapshot")
		}

		runtimeState, err := BuildRuntimeState(snapshot)
		if err != nil {
			return fmt.Errorf("build runtime state: %w", err)
		}

		runtimeConfig.Store(runtimeState)

		current := runtimeConfig.Load()

		log.Printf(
			"runtime config updated: %d routes, %d backends",
			len(current.Config.Routes),
			len(current.Config.Backends),
		)

		fmt.Printf(
			"received config: %d services, %d routes\n",
			len(snapshot.GetServices()),
			len(snapshot.GetRoutes()),
		)
	}
}

func (c *ControlClient) Close() error {
	return c.conn.Close()
}
