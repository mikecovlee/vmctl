package vm

import (
	"context"
	"errors"
	"testing"
)

func TestPauseRejectsCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for _, resume := range []bool{false, true} {
		if err := (Manager{}).Pause(ctx, "demo", resume); !errors.Is(err, context.Canceled) {
			t.Fatalf("expected cancellation, got %v", err)
		}
	}
}
