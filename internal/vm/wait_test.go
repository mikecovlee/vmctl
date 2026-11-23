package vm

import (
	"context"
	"errors"
	"testing"
)

func TestWaitInstallHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := (Manager{}).WaitInstall(ctx, "demo"); !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v, want canceled", err)
	}
}
