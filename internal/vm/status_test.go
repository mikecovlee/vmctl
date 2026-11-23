package vm

import (
	"context"
	"errors"
	"testing"
)

func TestStatusHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := (Manager{}).Status(ctx, "demo"); !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v", err)
	}
}

func TestDeleteRequiresConfirmation(t *testing.T) {
	if err := (Manager{}).Delete(context.Background(), "demo", false); !errors.Is(err, ErrConfirmation) {
		t.Fatalf("got %v", err)
	}
}
