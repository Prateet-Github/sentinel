package dataplane

import (
	"context"
	"testing"
	"time"
)

func TestRuntimeConfigWaitReadyBlocksUntilStore(t *testing.T) {
	runtimeConfig := NewRuntimeConfig()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ready := make(chan struct{})

	go func() {
		_ = runtimeConfig.WaitReady(ctx)
		close(ready)
	}()

	select {
	case <-ready:
		t.Fatal("WaitReady returned before Store")
	case <-time.After(50 * time.Millisecond):
		// expected: still blocked
	}

	runtimeConfig.Store(&RuntimeState{})

	select {
	case <-ready:
		// expected
	case <-time.After(time.Second):
		t.Fatal("WaitReady did not return after Store")
	}
}

func TestRuntimeConfigWaitReadyUnblocksAfterStore(t *testing.T) {
	runtimeConfig := NewRuntimeConfig()

	ctx := context.Background()

	done := make(chan struct{})

	go func() {
		_ = runtimeConfig.WaitReady(ctx)
		close(done)
	}()

	runtimeConfig.Store(&RuntimeState{})

	select {
	case <-done:
		// expected
	case <-time.After(time.Second):
		t.Fatal("WaitReady did not unblock after Store")
	}
}

func TestRuntimeConfigWaitReadyReturnsOnContextCancellation(t *testing.T) {
	runtimeConfig := NewRuntimeConfig()

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)

	go func() {
		done <- runtimeConfig.WaitReady(ctx)
	}()

	cancel()

	select {
	case err := <-done:
		if err != context.Canceled {
			t.Fatalf(
				"expected context.Canceled, got %v",
				err,
			)
		}

	case <-time.After(time.Second):
		t.Fatal(
			"WaitReady did not return after context cancellation",
		)
	}
}

func TestRuntimeConfigMultipleStoresDoNotPanic(t *testing.T) {
	runtimeConfig := NewRuntimeConfig()

	first := &RuntimeState{}
	second := &RuntimeState{}
	third := &RuntimeState{}

	runtimeConfig.Store(first)
	runtimeConfig.Store(second)
	runtimeConfig.Store(third)

	if got := runtimeConfig.Load(); got != third {
		t.Fatal("expected latest runtime state to be stored")
	}

	if err := runtimeConfig.WaitReady(context.Background()); err != nil {
		t.Fatalf("WaitReady returned error: %v", err)
	}
}

func TestRuntimeConfigLastRuntimeStateRemainsAvailable(t *testing.T) {
	runtimeConfig := NewRuntimeConfig()

	first := &RuntimeState{}
	second := &RuntimeState{}

	runtimeConfig.Store(first)

	if got := runtimeConfig.Load(); got != first {
		t.Fatal("expected first runtime state")
	}

	runtimeConfig.Store(second)

	if got := runtimeConfig.Load(); got != second {
		t.Fatal("expected second runtime state")
	}

	if err := runtimeConfig.WaitReady(context.Background()); err != nil {
		t.Fatalf("WaitReady returned error: %v", err)
	}

	if got := runtimeConfig.Load(); got != second {
		t.Fatal("expected latest runtime state to remain available")
	}
}
