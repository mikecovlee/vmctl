package vm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"vmctl/internal/state"
)

var snapshotName = regexp.MustCompile(`^[a-z][a-z0-9-]{0,47}$`)

// Snapshot copies persistent guest state only after QEMU has released the instance.
func (m Manager) Snapshot(ctx context.Context, name, label string) (err error) {
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
		c, err := m.Docker.InspectInstance(ctx, i.Directory)
		if err != nil {
			return err
		}
		if c.State.Running {
			return fmt.Errorf("%w: stop the instance before snapshot", ErrState)
		}
	}
	root, err := snapshotRoot(i.Directory)
	if err != nil {
		return err
	}
	target := filepath.Join(root, label)
	if _, err = os.Lstat(target); !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("%w: snapshot already exists", ErrExists)
	}
	tmp, err := os.MkdirTemp(root, ".pending-")
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			err = errors.Join(err, os.RemoveAll(tmp))
		}
	}()
	for _, file := range []string{"disk.qcow2", "nvram.fd", "tpm"} {
		source := filepath.Join(i.Directory, file)
		if _, e := os.Lstat(source); errors.Is(e, os.ErrNotExist) && file != "disk.qcow2" {
			continue
		} else if e != nil {
			return e
		}
		// cp preserves sparse disks and copies TPM state without traversing host paths.
		if e := copySnapshotPath(ctx, source, filepath.Join(tmp, file)); e != nil {
			return e
		}
	}
	data, err := json.Marshal(i.Record)
	if err != nil {
		return err
	}
	if err = state.Write(filepath.Join(tmp, "state.json"), data); err != nil {
		return err
	}
	return os.Rename(tmp, target)
}
