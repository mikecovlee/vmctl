package cli

import (
	"context"
	"errors"
	"testing"

	"vmctl/internal/vm"
)

func TestCheckpointRejectsNonpositiveTimeout(t *testing.T) {
	for _, action := range []string{"checkpoint", "restore-checkpoint"} {
		for _, duration := range []string{"0s", "-1s"} {
			t.Run(action+duration, func(t *testing.T) {
				command := New()
				command.SetArgs([]string{"--state-dir", t.TempDir(), action, "demo", "saved", "--yes", "--timeout", duration})
				if err := command.ExecuteContext(context.Background()); !errors.Is(err, vm.ErrState) {
					t.Fatalf("expected invalid timeout, got %v", err)
				}
			})
		}
	}
}
