package vm

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func snapshotRoot(instance string) (string, error) {
	root := filepath.Join(instance, "snapshots")
	if err := os.Mkdir(root, 0700); err != nil && !errors.Is(err, os.ErrExist) {
		return "", err
	}
	info, err := os.Lstat(root)
	if err != nil {
		return "", err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return "", fmt.Errorf("%w: unsafe snapshot directory", ErrState)
	}
	return root, nil
}

func (m Manager) Snapshots(ctx context.Context, name string) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	i, err := m.Load(name)
	if err != nil {
		return nil, err
	}
	root, err := snapshotRoot(i.Directory)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	result := []string{}
	for _, e := range entries {
		if e.IsDir() && snapshotName.MatchString(e.Name()) {
			result = append(result, e.Name())
		}
	}
	return result, nil
}
