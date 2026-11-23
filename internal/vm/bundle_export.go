package vm

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func (m Manager) Export(ctx context.Context, name, output string) (err error) {
	if err := ctx.Err(); err != nil {
		return err
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
			return fmt.Errorf("%w: stop instance before export", ErrState)
		}
	}
	staging, err := os.MkdirTemp("", "vmctl-export-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(staging)
	for _, file := range []string{"disk.qcow2", "nvram.fd", "tpm", "state.json", "spec.json", "credentials.json", "ssh-key", "ssh-key.pub"} {
		src := filepath.Join(i.Directory, file)
		if _, e := os.Lstat(src); errors.Is(e, os.ErrNotExist) && file != "disk.qcow2" && file != "state.json" && file != "spec.json" && file != "credentials.json" {
			continue
		} else if e != nil {
			return e
		}
		if e := copySnapshotPath(ctx, src, filepath.Join(staging, file)); e != nil {
			return e
		}
	}
	if err := writeManifest(ctx, staging); err != nil {
		return err
	}
	f, err := os.OpenFile(output, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, f.Close())
		if err != nil {
			err = errors.Join(err, os.Remove(output))
		}
	}()
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)
	walkErr := filepath.WalkDir(staging, func(path string, d os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if e = ctx.Err(); e != nil {
			return e
		}
		if d.IsDir() {
			return nil
		}
		info, e := d.Info()
		if e != nil {
			return e
		}
		rel, e := filepath.Rel(staging, path)
		if e != nil {
			return e
		}
		if e = tw.WriteHeader(&tar.Header{Name: filepath.ToSlash(rel), Mode: 0600, Size: info.Size()}); e != nil {
			return e
		}
		in, e := os.Open(path)
		if e != nil {
			return e
		}
		_, e = io.Copy(tw, in)
		return errors.Join(e, in.Close())
	})
	return errors.Join(walkErr, tw.Close(), gz.Close(), f.Sync())
}
