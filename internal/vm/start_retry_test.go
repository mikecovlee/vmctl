package vm

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
	"vmctl/internal/docker"
	"vmctl/internal/state"
)

func TestReviewFailedLaunchLeavesRetryableState(t *testing.T) {
	root := t.TempDir()
	script := filepath.Join(root, "docker")
	if err := os.WriteFile(script, []byte("#!/bin/sh\nexit 0\n"), 0700); err != nil {
		t.Fatal(err)
	}
	m := Manager{Store: state.Store{Root: filepath.Join(root, "state")}, Docker: docker.Client{Binary: script}}
	if err := m.Store.Save(state.Record{Name: "demo", Phase: state.Created}); err != nil {
		t.Fatal(err)
	}
	dir, err := m.Store.Path("demo")
	if err != nil {
		t.Fatal(err)
	}
	for name, data := range map[string]string{"spec.json": `{"name":"demo","install":{"mode":"manual"},"console":{"vncBind":"127.0.0.1","vncPort":0}}`, "credentials.json": `{"vncPassword":"example1"}`} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	if err := m.Start(ctx, "demo", true); err == nil {
		t.Fatal("missing QMP accepted")
	}
	record, err := m.Store.Load("demo")
	if err != nil {
		t.Fatal(err)
	}
	if record.Phase != state.Created {
		t.Fatalf("failed launch leaves non-retryable phase %s", record.Phase)
	}
}
