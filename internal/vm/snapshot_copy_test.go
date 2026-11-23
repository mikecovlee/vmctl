package vm

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"testing"
)

func TestSnapshotCopyExcludesRuntimeSocket(t *testing.T) {
	source := filepath.Join(t.TempDir(), "tpm")
	if err := os.Mkdir(source, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "state"), []byte("persistent"), 0600); err != nil {
		t.Fatal(err)
	}
	socket, err := net.Listen("unix", filepath.Join(source, "socket"))
	if err != nil {
		t.Fatal(err)
	}
	defer socket.Close()
	target := filepath.Join(t.TempDir(), "tpm")
	if err := copySnapshotPath(context.Background(), source, target); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(target, "state"))
	if err != nil || string(data) != "persistent" {
		t.Fatalf("state: %q %v", data, err)
	}
	if _, err := os.Lstat(filepath.Join(target, "socket")); !os.IsNotExist(err) {
		t.Fatalf("socket copied: %v", err)
	}
}

func TestSnapshotCopyRejectsSymlink(t *testing.T) {
	source := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(t.TempDir(), source); err != nil {
		t.Fatal(err)
	}
	if err := copySnapshotPath(context.Background(), source, filepath.Join(t.TempDir(), "copy")); err == nil {
		t.Fatal("symlink accepted")
	}
}
