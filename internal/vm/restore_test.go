package vm

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"vmctl/internal/docker"
	"vmctl/internal/state"
)

func TestRestoreRequiresConfirmation(t *testing.T) {
	err := (Manager{}).Restore(context.Background(), "demo", "baseline", false)
	if !errors.Is(err, ErrConfirmation) {
		t.Fatalf("expected confirmation error, got %v", err)
	}
}

func TestRestoreRecoversAllPersistentComponents(t *testing.T) {
	m := Manager{Store: state.Store{Root: t.TempDir()}, Docker: docker.Client{Binary: "/bin/true"}}
	if err := m.Store.Save(state.Record{Name: "demo", Phase: state.Verified, MAC: "52:54:00:01:02:03"}); err != nil {
		t.Fatal(err)
	}
	dir, err := m.Store.Path("demo")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "spec.json"), []byte(`{"name":"demo"}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "tpm"), 0700); err != nil {
		t.Fatal(err)
	}
	files := []string{"disk.qcow2", "nvram.fd", "tpm/state"}
	for _, file := range files {
		if err := os.WriteFile(filepath.Join(dir, file), []byte("original"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := m.Snapshot(context.Background(), "demo", "baseline"); err != nil {
		t.Fatal(err)
	}
	// Required firmware must be checked before replacing any live file.
	if err := os.Remove(filepath.Join(dir, "snapshots/baseline/nvram.fd")); err != nil {
		t.Fatal(err)
	}
	i, err := os.ReadFile(filepath.Join(dir, "spec.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "spec.json"), []byte(`{"name":"demo","machine":{"firmware":"uefi","tpm":true}}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := m.Restore(context.Background(), "demo", "baseline", true); err == nil {
		t.Fatal("missing firmware accepted")
	}
	if _, err := os.Stat(filepath.Join(dir, "nvram.fd")); err != nil {
		t.Fatal("live firmware lost")
	}
	if err := os.WriteFile(filepath.Join(dir, "spec.json"), i, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "snapshots/baseline/nvram.fd"), []byte("original"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		if err := os.WriteFile(filepath.Join(dir, file), []byte("changed"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := m.Restore(context.Background(), "demo", "baseline", true); err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		data, err := os.ReadFile(filepath.Join(dir, file))
		if err != nil || string(data) != "original" {
			t.Fatalf("%s: %q %v", file, data, err)
		}
	}
}

func TestRestoreRejectsCorruptRecordBeforeChangingDisk(t *testing.T) {
	m := Manager{Store: state.Store{Root: t.TempDir()}, Docker: docker.Client{Binary: "/bin/true"}}
	if err := m.Store.Save(state.Record{Name: "demo", Phase: state.Verified}); err != nil {
		t.Fatal(err)
	}
	dir, err := m.Store.Path("demo")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "spec.json"), []byte(`{"name":"demo"}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "disk.qcow2"), []byte("live"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := m.Snapshot(context.Background(), "demo", "broken"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "snapshots/broken/state.json"), []byte("invalid"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := m.Restore(context.Background(), "demo", "broken", true); err == nil {
		t.Fatal("corrupt snapshot accepted")
	}
	data, err := os.ReadFile(filepath.Join(dir, "disk.qcow2"))
	if err != nil || string(data) != "live" {
		t.Fatalf("disk changed: %q %v", data, err)
	}
}
