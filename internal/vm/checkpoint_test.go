package vm

import (
	"context"
	"errors"
	"testing"
)

func TestCheckpointHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := (Manager{}).Checkpoint(ctx, "demo", "saved", false); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation, got %v", err)
	}
}
