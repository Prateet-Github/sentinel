package dataplane

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"

	controlv1 "github.com/Prateet-Github/sentinel/proto"
	"google.golang.org/grpc"
)

type testControlServer struct {
	controlv1.UnimplementedSentinelControlServer

	mu sync.RWMutex

	snapshot *controlv1.ConfigSnapshot

	// When closeStream is closed, the active StreamConfig
	// RPC returns and the client must reconnect.
	closeStream chan struct{}
}

func (s *testControlServer) StreamConfig(
	stream controlv1.SentinelControl_StreamConfigServer,
) error {
	req, err := stream.Recv()
	if err != nil {
		return err
	}

	if req.GetNodeId() != "dp-test" {
		return nil
	}

	s.mu.RLock()
	snapshot := s.snapshot
	closeStream := s.closeStream
	s.mu.RUnlock()

	if err := stream.Send(&controlv1.ConfigResponse{
		Snapshot: snapshot,
	}); err != nil {
		return err
	}

	if closeStream == nil {
		<-stream.Context().Done()
		return stream.Context().Err()
	}

	select {
	case <-closeStream:
		return nil

	case <-stream.Context().Done():
		return stream.Context().Err()
	}
}

func startTestControlServer(
	t *testing.T,
	serverImpl *testControlServer,
) (string, func()) {
	t.Helper()

	listener, err := net.Listen(
		"tcp",
		"127.0.0.1:0",
	)
	if err != nil {
		t.Fatal(err)
	}

	server := grpc.NewServer()

	controlv1.RegisterSentinelControlServer(
		server,
		serverImpl,
	)

	go func() {
		_ = server.Serve(listener)
	}()

	stop := func() {
		server.Stop()
		_ = listener.Close()
	}

	return listener.Addr().String(), stop
}

func TestControlClientReceivesSnapshot(t *testing.T) {
	snapshot := &controlv1.ConfigSnapshot{
		Services: []*controlv1.Service{
			{
				Name: "users",
				Backends: []*controlv1.Backend{
					{
						Name:            "users-1",
						Url:             "http://127.0.0.1:9000",
						HealthCheckPath: "/health",
					},
				},
			},
		},
	}

	address, stop := startTestControlServer(
		t,
		&testControlServer{
			snapshot: snapshot,
		},
	)
	defer stop()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	runtimeConfig := NewRuntimeConfig()

	client, err := NewControlClient(
		ctx,
		address,
	)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	done := make(chan error, 1)

	go func() {
		done <- client.StreamConfig(
			ctx,
			"dp-test",
			runtimeConfig,
		)
	}()

	select {
	case <-runtimeConfig.ready:

	case <-time.After(3 * time.Second):
		t.Fatal(
			"runtime config was not initialized",
		)
	}

	state := runtimeConfig.Load()

	if state == nil {
		t.Fatal("expected runtime state")
	}

	if len(state.Config.Backends) != 1 {
		t.Fatalf(
			"expected 1 backend, got %d",
			len(state.Config.Backends),
		)
	}

	cancel()

	select {
	case <-done:

	case <-time.After(2 * time.Second):
		t.Fatal(
			"control client did not stop",
		)
	}
}

func TestControlClientReconnectsAfterControlPlaneRestart(
	t *testing.T,
) {
	snapshot1 := &controlv1.ConfigSnapshot{
		Services: []*controlv1.Service{
			{
				Name: "users",
				Backends: []*controlv1.Backend{
					{
						Name:            "users-1",
						Url:             "http://127.0.0.1:9000",
						HealthCheckPath: "/health",
					},
				},
			},
		},
	}

	snapshot2 := &controlv1.ConfigSnapshot{
		Services: []*controlv1.Service{
			{
				Name: "users",
				Backends: []*controlv1.Backend{
					{
						Name:            "users-1",
						Url:             "http://127.0.0.1:9000",
						HealthCheckPath: "/health",
					},
					{
						Name:            "users-2",
						Url:             "http://127.0.0.1:9001",
						HealthCheckPath: "/health",
					},
				},
			},
		},
	}

	// Start CP #1.
	listener, err := net.Listen(
		"tcp",
		"127.0.0.1:0",
	)
	if err != nil {
		t.Fatal(err)
	}

	address := listener.Addr().String()

	server1 := grpc.NewServer()

	controlv1.RegisterSentinelControlServer(
		server1,
		&testControlServer{
			snapshot: snapshot1,
		},
	)

	go func() {
		_ = server1.Serve(listener)
	}()

	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	defer cancel()

	runtimeConfig := NewRuntimeConfig()

	client, err := NewControlClient(
		ctx,
		address,
	)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	done := make(chan error, 1)

	go func() {
		done <- client.StreamConfig(
			ctx,
			"dp-test",
			runtimeConfig,
		)
	}()

	// Wait for the initial configuration.
	select {
	case <-runtimeConfig.ready:

	case <-time.After(3 * time.Second):
		t.Fatal(
			"initial runtime config was not received",
		)
	}

	state := runtimeConfig.Load()

	if state == nil {
		t.Fatal("expected runtime state")
	}

	if len(state.Config.Backends) != 1 {
		t.Fatalf(
			"expected 1 backend initially, got %d",
			len(state.Config.Backends),
		)
	}

	/*
		Stop CP #1 completely.

		This closes the listener, so the next reconnect
		attempts will receive connection refused.
	*/
	server1.Stop()
	_ = listener.Close()

	/*
		Wait until the client has observed the broken
		connection and entered its reconnect loop.
	*/
	time.Sleep(150 * time.Millisecond)

	/*
		Start CP #2 on the SAME address.

		The existing ControlClient still has:
		    address = 127.0.0.1:<same-port>

		so it can reconnect without changing anything
		on the client side.
	*/
	listener2, err := net.Listen(
		"tcp",
		address,
	)
	if err != nil {
		t.Fatalf(
			"failed to restart control plane on %s: %v",
			address,
			err,
		)
	}

	server2 := grpc.NewServer()

	controlv1.RegisterSentinelControlServer(
		server2,
		&testControlServer{
			snapshot: snapshot2,
		},
	)

	go func() {
		_ = server2.Serve(listener2)
	}()

	/*
		Wait for the existing ControlClient to reconnect
		and receive snapshot2.

		The reconnect sequence can be:

		    1s
		    2s
		    4s

		so give it enough time.
	*/
	deadline := time.After(10 * time.Second)

	for {
		state = runtimeConfig.Load()

		if state != nil &&
			len(state.Config.Backends) == 2 {
			break
		}

		select {
		case <-deadline:
			t.Fatal(
				"data plane did not receive updated snapshot after reconnect",
			)

		case <-time.After(50 * time.Millisecond):
		}
	}

	// Verify the new configuration.
	if state.Config.Backends[0].Name != "users-1" {
		t.Fatalf(
			"expected first backend users-1, got %s",
			state.Config.Backends[0].Name,
		)
	}

	if state.Config.Backends[1].Name != "users-2" {
		t.Fatalf(
			"expected second backend users-2, got %s",
			state.Config.Backends[1].Name,
		)
	}

	server2.Stop()
	_ = listener2.Close()

	cancel()

	select {
	case <-done:

	case <-time.After(2 * time.Second):
		t.Fatal(
			"control client did not stop",
		)
	}
}

func TestControlClientKeepsLastRuntimeStateDuringControlPlaneOutage(
	t *testing.T,
) {
	snapshot := &controlv1.ConfigSnapshot{
		Services: []*controlv1.Service{
			{
				Name: "users",
				Backends: []*controlv1.Backend{
					{
						Name:            "users-1",
						Url:             "http://127.0.0.1:9000",
						HealthCheckPath: "/health",
					},
				},
			},
		},
	}

	listener, err := net.Listen(
		"tcp",
		"127.0.0.1:0",
	)
	if err != nil {
		t.Fatal(err)
	}

	address := listener.Addr().String()

	server := grpc.NewServer()

	controlv1.RegisterSentinelControlServer(
		server,
		&testControlServer{
			snapshot: snapshot,
		},
	)

	go func() {
		_ = server.Serve(listener)
	}()

	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	defer cancel()

	runtimeConfig := NewRuntimeConfig()

	client, err := NewControlClient(
		ctx,
		address,
	)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	done := make(chan error, 1)

	go func() {
		done <- client.StreamConfig(
			ctx,
			"dp-test",
			runtimeConfig,
		)
	}()

	// Wait for initial configuration.
	select {
	case <-runtimeConfig.ready:

	case <-time.After(3 * time.Second):
		t.Fatal(
			"initial runtime config was not received",
		)
	}

	// Capture the active runtime state.
	initialState := runtimeConfig.Load()

	if initialState == nil {
		t.Fatal("expected runtime state")
	}

	if len(initialState.Config.Backends) != 1 {
		t.Fatalf(
			"expected 1 backend, got %d",
			len(initialState.Config.Backends),
		)
	}

	// Simulate complete CP outage.
	server.Stop()
	_ = listener.Close()

	// Give the client time to observe the outage.
	time.Sleep(200 * time.Millisecond)

	// Runtime state must still exist.
	currentState := runtimeConfig.Load()

	if currentState == nil {
		t.Fatal(
			"runtime state was lost during control plane outage",
		)
	}

	// It must still contain the previous configuration.
	if len(currentState.Config.Backends) != 1 {
		t.Fatalf(
			"expected last-known configuration to remain active, got %d backends",
			len(currentState.Config.Backends),
		)
	}

	if currentState.Config.Backends[0].Name != "users-1" {
		t.Fatalf(
			"expected users-1 to remain active, got %s",
			currentState.Config.Backends[0].Name,
		)
	}

	cancel()

	select {
	case <-done:

	case <-time.After(2 * time.Second):
		t.Fatal(
			"control client did not stop",
		)
	}
}
