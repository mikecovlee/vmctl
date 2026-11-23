package docker

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRunPreservesArguments(t *testing.T) {
	// Given an executable that echoes each argument without shell expansion.
	c := Client{Binary: "/bin/echo", Timeout: time.Second}
	// When shell metacharacters occur in an argument.
	out, err := c.Run(context.Background(), "$(touch should-not-exist)")
	// Then they remain literal data.
	if err != nil || string(out) != "$(touch should-not-exist)\n" {
		t.Fatalf("%q: %v", out, err)
	}
}

func TestRunHonorsCancellation(t *testing.T) {
	// Given an already cancelled operation.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	// When an external command is requested.
	_, err := (Client{Binary: "/bin/sleep", Timeout: time.Second}).Run(ctx, "30")
	// Then cancellation is exposed to the caller.
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v", err)
	}
}
