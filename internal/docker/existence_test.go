package docker

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestExistsDistinguishesAbsentContainer(t *testing.T) {
	for _, output := range []string{"", "container-id\n"} {
		t.Run(output, func(t *testing.T) {
			binary := filepath.Join(t.TempDir(), "docker")
			if err := os.WriteFile(binary, []byte("#!/bin/sh\nprintf '"+output+"'\n"), 0700); err != nil {
				t.Fatal(err)
			}
			found, err := (Client{Binary: binary}).Exists(context.Background(), "/private/demo")
			if err != nil {
				t.Fatal(err)
			}
			if found != (output != "") {
				t.Fatalf("exists=%v for %q", found, output)
			}
		})
	}
}
