package docker

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestInspectRejectsForeignInstance(t *testing.T) {
	// Given a managed container belonging to a different state directory.
	binary := filepath.Join(t.TempDir(), "docker")
	script := "#!/bin/sh\nprintf '%s' '[{\"Id\":\"abc\",\"Config\":{\"Labels\":{\"io.vmctl.managed\":\"true\",\"io.vmctl.instance\":\"/other/demo\"}}}]'\n"
	if err := os.WriteFile(binary, []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	// When the caller inspects it for its own instance directory.
	_, err := (Client{Binary: binary}).InspectInstance(context.Background(), "/private/demo")
	// Then ownership is rejected before any mutation.
	if !errors.Is(err, ErrOwnership) {
		t.Fatalf("got %v, want ownership rejection", err)
	}
}
