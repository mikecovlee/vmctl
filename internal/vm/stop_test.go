package vm

import (
	"context"
	"errors"
	"testing"
)

func TestStopHonorsCanceledContext(t *testing.T) {
	// Given a canceled operation, no Docker command should be needed.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	// When stopping an instance.
	err := (Manager{}).Stop(ctx, "demo", false)
	// Then cancellation is propagated rather than forcing a shutdown.
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v", err)
	}
}
