package vm

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"vmctl/internal/config"
	"vmctl/internal/docker"
	"vmctl/internal/state"
)

func TestLoadPreservesCreatedConfiguration(t *testing.T) {
	iso := filepath.Join(t.TempDir(), "installer.iso")
	if err := os.WriteFile(iso, []byte("fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	m := Manager{Store: state.Store{Root: t.TempDir()}, Docker: docker.Client{Binary: "/bin/true"}}
	spec := config.Spec{Name: "demo", Media: config.Media{ISO: iso}}
	if _, err := m.Create(context.Background(), spec); err != nil {
		t.Fatal(err)
	}
	loaded, err := m.Load("demo")
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Spec.Media.ISO != iso || loaded.Record.Phase != state.Created {
		t.Fatalf("unexpected instance: %+v", loaded)
	}
}
