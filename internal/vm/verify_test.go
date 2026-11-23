package vm

import (
	"context"
	"errors"
	"testing"
)

func TestVerifyHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := (Manager{}).Verify(ctx, "demo"); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation, got %v", err)
	}
}
