package iap

import (
	"context"
	"testing"
	"time"
)

// Cancelling the context passed to Listen must close the listener so the
// accept goroutine terminates rather than leaking.
func TestListenClosesListenerOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	ln, err := Listen(ctx, TunnelConfig{}, 0)
	if err != nil {
		t.Fatalf("Listen: %v", err)
	}
	defer ln.Close()

	cancel()

	// Once the context is cancelled, the listener is closed and Accept returns.
	done := make(chan error, 1)
	go func() {
		_, e := ln.Accept()
		done <- e
	}()

	select {
	case e := <-done:
		if e == nil {
			t.Error("expected Accept to return an error after context cancel")
		}
	case <-time.After(2 * time.Second):
		t.Error("Accept did not return after context cancel; listener not closed")
	}
}
