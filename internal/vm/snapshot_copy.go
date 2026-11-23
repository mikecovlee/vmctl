package vm

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
)

func copySnapshotPath(ctx context.Context, source, target string) error {
	return filepath.WalkDir(source, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Type()&fs.ModeSymlink != 0 {
			return fmt.Errorf("%w: symlink in snapshot source", ErrState)
		}
		if d.Type()&fs.ModeSocket != 0 && d.Name() == "socket" {
			return nil
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		destination := filepath.Join(target, relative)
		if d.IsDir() {
			return os.Mkdir(destination, 0700)
		}
		if !d.Type().IsRegular() {
			return fmt.Errorf("%w: special file in snapshot source", ErrState)
		}
		cmd := exec.CommandContext(ctx, "cp", "--sparse=always", "--", path, destination)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("copy snapshot: %w: %s", err, out)
		}
		return os.Chmod(destination, 0600)
	})
}
