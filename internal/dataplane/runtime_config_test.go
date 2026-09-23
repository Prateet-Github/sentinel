package dataplane

import (
	"testing"
	"time"
)

func TestRuntimeConfigWaitReadyBlocksUntilStore(t *testing.T) {
	runtimeConfig := NewRuntimeConfig()

	ready := make(chan struct{})

	go func() {
		runtimeConfig.WaitReady()
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

	done := make(chan struct{})

	go func() {
		runtimeConfig.WaitReady()
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

	runtimeConfig.WaitReady()
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

	runtimeConfig.WaitReady()

	if got := runtimeConfig.Load(); got != second {
		t.Fatal("expected latest runtime state to remain available")
	}
}
