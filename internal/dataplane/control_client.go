package dataplane

import (
	"context"
	"fmt"
	"log"
	"time"

	controlv1 "github.com/Prateet-Github/sentinel/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ControlClient struct {
	client controlv1.SentinelControlClient
	conn   *grpc.ClientConn
}

const (
	initialReconnectDelay = 1 * time.Second
	maxReconnectDelay     = 10 * time.Second
)

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
	reconnectDelay := initialReconnectDelay

	for {
		connected, err := c.streamConfigOnce(
			ctx,
			nodeID,
			runtimeConfig,
		)

		if ctx.Err() != nil {
			return ctx.Err()
		}

		log.Printf(
			"control plane stream disconnected: %v",
			err,
		)

		if connected {
			reconnectDelay = initialReconnectDelay
		}

		log.Printf(
			"reconnecting in %s",
			reconnectDelay,
		)

		timer := time.NewTimer(reconnectDelay)

		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return ctx.Err()

		case <-timer.C:
		}

		if reconnectDelay < maxReconnectDelay {
			reconnectDelay *= 2

			if reconnectDelay > maxReconnectDelay {
				reconnectDelay = maxReconnectDelay
			}
		}
	}
}

func (c *ControlClient) streamConfigOnce(
	ctx context.Context,
	nodeID string,
	runtimeConfig *RuntimeConfig,
) (bool, error) {
	stream, err := c.client.StreamConfig(ctx)
	if err != nil {
		return false, err
	}

	if err := stream.Send(&controlv1.ConfigRequest{
		NodeId: nodeID,
	}); err != nil {
		return false, err
	}

	// connection/stream was successfully established
	connected := true

	for {
		response, err := stream.Recv()
		if err != nil {
			return connected, err
		}

		snapshot := response.GetSnapshot()
		if snapshot == nil {
			return connected, fmt.Errorf(
				"received empty config snapshot",
			)
		}

		runtimeState, err := BuildRuntimeState(snapshot)
		if err != nil {
			return connected, fmt.Errorf(
				"build runtime state: %w",
				err,
			)
		}

		runtimeConfig.Store(runtimeState)

		current := runtimeConfig.Load()

		log.Printf(
			"runtime config updated: %d routes, %d backends",
			len(current.Config.Routes),
			len(current.Config.Backends),
		)

		log.Printf(
			"received config: %d services, %d routes",
			len(snapshot.GetServices()),
			len(snapshot.GetRoutes()),
		)
	}
}

func (c *ControlClient) Close() error {
	return c.conn.Close()
}
