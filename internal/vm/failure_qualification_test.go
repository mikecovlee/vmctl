package vm

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"vmctl/internal/config"
	"vmctl/internal/docker"
	"vmctl/internal/state"
)

func TestCreateDiskFailureLeavesNoPublishedInstance(t *testing.T) {
	// Given a valid configuration and a disk allocator reporting exhausted space.
	root := t.TempDir()
	iso := filepath.Join(root, "install.iso")
	if err := os.WriteFile(iso, []byte("media"), 0600); err != nil {
		t.Fatal(err)
	}
	spec, err := config.Parse(bytes.NewBufferString(`{"apiVersion":"v1alpha1","name":"source","resources":{"cpu":1,"memoryMiB":512,"diskGiB":2},"media":{"iso":"` + iso + `"}}`))
	if err != nil {
		t.Fatal(err)
	}
	allocator := filepath.Join(root, "allocator")
	if err := os.WriteFile(allocator, []byte("#!/bin/sh\nprintf 'No space left on device\\n' >&2\nexit 1\n"), 0700); err != nil {
		t.Fatal(err)
	}
	m := Manager{Store: state.Store{Root: filepath.Join(root, "state")}, Docker: docker.Client{Binary: allocator}}
	// When allocation fails after private instance metadata has been written.
	if _, err := m.Create(context.Background(), spec); err == nil {
		t.Fatal("failed allocator accepted")
	}
	// Then no instance is published and the input media is retained.
	if _, err := os.Stat(filepath.Join(m.Store.Root, "source")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("partial instance remains: %v", err)
	}
	if data, err := os.ReadFile(iso); err != nil || string(data) != "media" {
		t.Fatalf("input changed: %q %v", data, err)
	}
}

func TestSnapshotCopyFailurePreservesSourceAndRemovesStaging(t *testing.T) {
	// Given an installed instance and a copy process reporting exhausted space.
	root := t.TempDir()
	m := Manager{Store: state.Store{Root: filepath.Join(root, "state")}, Docker: docker.Client{Binary: "/bin/true"}}
	if err := m.Store.Save(state.Record{Name: "demo", Phase: state.Installed}); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(m.Store.Root, "demo")
	for name, data := range map[string]string{"spec.json": `{"name":"demo"}`, "credentials.json": `{"vncPassword":"example1"}`, "disk.qcow2": "original-disk"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	tools := filepath.Join(root, "tools")
	if err := os.Mkdir(tools, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tools, "cp"), []byte("#!/bin/sh\nprintf 'No space left on device\\n' >&2\nexit 1\n"), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", tools)
	// When the persistent-state copy fails before publication.
	if err := m.Snapshot(context.Background(), "demo", "baseline"); err == nil {
		t.Fatal("failed copy accepted")
	}
	// Then the original disk is unchanged and neither a snapshot nor staging survives.
	data, err := os.ReadFile(filepath.Join(dir, "disk.qcow2"))
	if err != nil || string(data) != "original-disk" {
		t.Fatalf("source changed: %q %v", data, err)
	}
	entries, err := os.ReadDir(filepath.Join(dir, "snapshots"))
	if err != nil || len(entries) != 0 {
		t.Fatalf("partial snapshot remains: %v %v", entries, err)
	}
}
