package vm

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSnapshotRootRejectsSymlink(t *testing.T) {
	dir := t.TempDir()
	if err := os.Symlink(t.TempDir(), filepath.Join(dir, "snapshots")); err != nil {
		t.Fatal(err)
	}
	if _, err := snapshotRoot(dir); err == nil {
		t.Fatal("accepted symlink snapshot root")
	}
}

func TestSnapshotRootCreatesPrivateDirectory(t *testing.T) {
	root, err := snapshotRoot(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(root)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0700 {
		t.Fatalf("permissions %o", info.Mode().Perm())
	}
}
