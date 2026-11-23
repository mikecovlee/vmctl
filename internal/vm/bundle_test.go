package vm

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"vmctl/internal/config"
	"vmctl/internal/docker"
	"vmctl/internal/state"
)

func TestImportRejectsTraversal(t *testing.T) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	tw := tar.NewWriter(gz)
	if err := tw.WriteHeader(&tar.Header{Name: "../escape", Mode: 0600, Size: 1}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write([]byte("x")); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	file := filepath.Join(dir, "bad.tgz")
	if err := os.WriteFile(file, b.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	if err := extractBundle(context.Background(), file, t.TempDir()); err == nil {
		t.Fatal("unsafe archive accepted")
	}
}

func TestBundleRoundTripPreservesIdentity(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	iso := filepath.Join(root, "install.iso")
	if err := os.WriteFile(iso, []byte("media"), 0600); err != nil {
		t.Fatal(err)
	}
	spec, err := config.Parse(bytes.NewBufferString(`{"apiVersion":"v1alpha1","name":"source","resources":{"cpu":1,"memoryMiB":512,"diskGiB":2},"media":{"iso":"` + iso + `"}}`))
	if err != nil {
		t.Fatal(err)
	}
	m := Manager{Store: state.Store{Root: filepath.Join(root, "instances")}, Docker: docker.Client{Binary: "/bin/true"}}
	record, err := m.Create(ctx, spec)
	if err != nil {
		t.Fatal(err)
	}
	dir, err := m.Store.Path("source")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "disk.qcow2"), []byte("disk-state"), 0600); err != nil {
		t.Fatal(err)
	}
	archive := filepath.Join(root, "export.tgz")
	probe := filepath.Join(root, "docker-probe")
	if err := os.WriteFile(probe, []byte("#!/bin/sh\ncase \" $* \" in *' info '*) printf '{\"format\":\"qcow2\"}';; esac\n"), 0700); err != nil {
		t.Fatal(err)
	}
	m.Docker.Binary = probe
	if err := m.Export(ctx, "source", archive); err != nil {
		t.Fatal(err)
	}
	got, err := m.Import(ctx, ImportOptions{Archive: archive, Name: "restored"})
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "restored" || got.MAC != record.MAC {
		t.Fatalf("identity mismatch: %+v", got)
	}
	data, err := os.ReadFile(filepath.Join(root, "instances/restored/disk.qcow2"))
	if err != nil || string(data) != "disk-state" {
		t.Fatalf("disk mismatch: %q %v", data, err)
	}
	if _, err := m.Import(ctx, ImportOptions{Archive: archive, Name: "restored"}); !errors.Is(err, ErrExists) {
		t.Fatalf("overwrite accepted: %v", err)
	}
}
