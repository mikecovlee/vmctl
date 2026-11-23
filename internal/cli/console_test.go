package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"vmctl/internal/config"
	"vmctl/internal/docker"
	"vmctl/internal/state"
	"vmctl/internal/vm"
)

func TestConsoleKeepsSecretPrivate(t *testing.T) {
	root := t.TempDir()
	iso := filepath.Join(t.TempDir(), "install.iso")
	if err := os.WriteFile(iso, []byte("fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	m := vm.Manager{Store: state.Store{Root: root}, Docker: docker.Client{Binary: "/bin/true"}}
	_, err := m.Create(context.Background(), config.Spec{Name: "demo", Media: config.Media{ISO: iso}, Console: config.Console{VNCBind: "127.0.0.1", VNCPort: 5909}})
	if err != nil {
		t.Fatal(err)
	}
	for _, reveal := range []bool{false, true} {
		cmd := New()
		var output bytes.Buffer
		cmd.SetOut(&output)
		args := []string{"--state-dir", root, "console", "demo"}
		if reveal {
			args = append(args, "--show-secret")
		}
		cmd.SetArgs(args)
		if err := cmd.Execute(); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(output.String(), "5909") {
			t.Fatal(output.String())
		}
		if strings.Contains(output.String(), "password") != reveal {
			t.Fatalf("unexpected secret visibility: %s", output.String())
		}
	}
}
