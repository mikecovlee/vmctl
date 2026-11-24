package vm

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"vmctl/internal/docker"
	"vmctl/internal/state"
)

func TestStartRejectsInterruptedRestoreBeforeLaunching(t *testing.T) {
	for _, pending := range []string{".restore-interrupted", ".rollback-interrupted"} {
		t.Run(pending, func(t *testing.T) {
			m := Manager{Store: state.Store{Root: t.TempDir()}, Docker: docker.Client{Binary: "/bin/true"}}
			if err := m.Store.Save(state.Record{Name: "demo", Phase: state.Installed}); err != nil {
				t.Fatal(err)
			}
			dir, err := m.Store.Path("demo")
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "spec.json"), []byte(`{"name":"demo"}`), 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.Mkdir(filepath.Join(dir, pending), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "credentials.json"), []byte(`{"vncPassword":"example1"}`), 0600); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()
			if err := m.Start(ctx, "demo", false); !errors.Is(err, ErrState) {
				t.Fatalf("interrupted restore must block boot: %v", err)
			}
		})
	}
}
