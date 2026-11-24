package vm

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"os"
	"path/filepath"
	"testing"
	"vmctl/internal/config"
	"vmctl/internal/docker"
)

func TestRequiredFirmwareMissingIsRejected(t *testing.T) {
	if err := validatePersistentFiles(t.TempDir(), config.Machine{Firmware: "uefi", TPM: true}); err == nil {
		t.Fatal("missing firmware accepted")
	}
}

func TestDiskValidationRejectsRawAndBackingFiles(t *testing.T) {
	for _, output := range []string{`{"format":"raw"}`, `{"format":"qcow2","backing-filename":"other.qcow2"}`, `{"format":"qcow2","format-specific":{"type":"qcow2","data":{"data-file":"/instance/external.raw"}}}`, `not json`} {
		path := filepath.Join(t.TempDir(), "docker")
		if err := os.WriteFile(path, []byte("#!/bin/sh\nprintf '%s' '"+output+"'\n"), 0700); err != nil {
			t.Fatal(err)
		}
		m := Manager{Docker: docker.Client{Binary: path}}
		if err := m.validateBundleDisk(context.Background(), t.TempDir()); err == nil {
			t.Fatalf("accepted %s", output)
		}
	}
}

func TestExtractionRejectsOversizedMetadataBeforeReadingPayload(t *testing.T) {
	var data bytes.Buffer
	gz := gzip.NewWriter(&data)
	tw := tar.NewWriter(gz)
	if err := tw.WriteHeader(&tar.Header{Name: "credentials.json", Mode: 0600, Size: 2 << 20}); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "oversized.tgz")
	if err := os.WriteFile(path, data.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	if err := extractBundle(context.Background(), path, t.TempDir()); err == nil {
		t.Fatal("oversized metadata accepted")
	}
}

func TestCredentialsRejectMalformedAndEmpty(t *testing.T) {
	for _, value := range []string{"not json", `{}`, `{"vncPassword":""}`, `{"vncPassword":"abcdefgh","unknown":1}`} {
		dir := t.TempDir()
		path := filepath.Join(dir, "credentials.json")
		if err := os.WriteFile(path, []byte(value), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := readCredentials(path); err == nil {
			t.Fatalf("accepted %q", value)
		}
	}
}

func TestCopyHonorsCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := copyContext(ctx, os.Stdout, os.Stdin); err == nil {
		t.Fatal("cancellation ignored")
	}
}
