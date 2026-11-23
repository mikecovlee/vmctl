package vm

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func (m Manager) Restore(ctx context.Context, name, label string, confirmed bool) (err error) {
	if !confirmed {
		return ErrConfirmation
	}
	if !snapshotName.MatchString(label) {
		return fmt.Errorf("%w: invalid snapshot name", ErrState)
	}
	lock, err := m.Store.Lock(name)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, lock.Close()) }()
	i, err := m.Load(name)
	if err != nil {
		return err
	}
	exists, err := m.Docker.Exists(ctx, i.Directory)
	if err != nil {
		return err
	}
	if exists {
		c, e := m.Docker.InspectInstance(ctx, i.Directory)
		if e != nil {
			return e
		}
		if c.State.Running {
			return fmt.Errorf("%w: stop the instance before restore", ErrState)
		}
	}
	root, err := snapshotRoot(i.Directory)
	if err != nil {
		return err
	}
	source := filepath.Join(root, label)
	info, err := os.Lstat(source)
	if err != nil {
		return err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return ErrState
	}
	staged, err := os.MkdirTemp(i.Directory, ".restore-")
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, os.RemoveAll(staged)) }()
	files := []string{"disk.qcow2", "nvram.fd", "tpm", "state.json"}
	for _, file := range files {
		path := filepath.Join(source, file)
		if _, e := os.Lstat(path); errors.Is(e, os.ErrNotExist) && (file == "nvram.fd" || file == "tpm") {
			continue
		} else if e != nil {
			return e
		}
		if e := copySnapshotPath(ctx, path, filepath.Join(staged, file)); e != nil {
			return e
		}
	}
	if err = validateSnapshotRecord(filepath.Join(staged, "state.json"), name); err != nil {
		return err
	}
	if err = ctx.Err(); err != nil {
		return err
	}
	backup, err := os.MkdirTemp(i.Directory, ".rollback-")
	if err != nil {
		return err
	}
	var moved, installed []string
	rollback := func(cause error) error {
		for _, file := range installed {
			cause = errors.Join(cause, os.RemoveAll(filepath.Join(i.Directory, file)))
		}
		for _, file := range moved {
			cause = errors.Join(cause, os.Rename(filepath.Join(backup, file), filepath.Join(i.Directory, file)))
		}
		return fmt.Errorf("restore failed (recovery directory %s): %w", backup, cause)
	}
	for _, file := range files {
		live := filepath.Join(i.Directory, file)
		if _, e := os.Lstat(live); e == nil {
			if e = os.Rename(live, filepath.Join(backup, file)); e != nil {
				return rollback(e)
			}
			moved = append(moved, file)
		} else if !errors.Is(e, os.ErrNotExist) {
			return rollback(e)
		}
		if _, e := os.Lstat(filepath.Join(staged, file)); e == nil {
			if e = os.Rename(filepath.Join(staged, file), live); e != nil {
				return rollback(e)
			}
			installed = append(installed, file)
		} else if !errors.Is(e, os.ErrNotExist) {
			return rollback(e)
		}
	}
	return os.RemoveAll(backup)
}
