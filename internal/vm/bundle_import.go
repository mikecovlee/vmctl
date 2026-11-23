package vm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"vmctl/internal/config"
	"vmctl/internal/state"
)

type ImportOptions struct {
	Archive string
	Name    string
	Config  string
}

func (m Manager) Import(ctx context.Context, opts ImportOptions) (record state.Record, err error) {
	if err := ctx.Err(); err != nil {
		return record, err
	}
	lock, err := m.Store.Lock(opts.Name)
	if err != nil {
		return record, err
	}
	defer func() { err = errors.Join(err, lock.Close()) }()
	target, err := m.Store.Path(opts.Name)
	if err != nil {
		return record, err
	}
	if _, e := os.Lstat(target); !errors.Is(e, os.ErrNotExist) {
		return record, ErrExists
	}
	tmp, err := os.MkdirTemp(filepath.Dir(target), ".import-")
	if err != nil {
		return record, err
	}
	defer os.RemoveAll(tmp)
	if err := extractBundle(ctx, opts.Archive, tmp); err != nil {
		return record, err
	}
	spec, err := config.Read(filepath.Join(tmp, "spec.json"))
	if err != nil {
		return record, err
	}
	if err := validateSnapshotRecord(filepath.Join(tmp, "state.json"), spec.Name); err != nil {
		return record, err
	}
	data, err := os.ReadFile(filepath.Join(tmp, "state.json"))
	if err != nil {
		return record, err
	}
	if err = json.Unmarshal(data, &record); err != nil {
		return record, err
	}
	if opts.Config != "" {
		override, e := config.Read(opts.Config)
		if e != nil {
			return record, e
		}
		if override.Machine != spec.Machine || override.Resources != spec.Resources {
			return record, fmt.Errorf("%w: import must preserve machine and resources", ErrState)
		}
		spec.Media = override.Media
		spec.Network = override.Network
		spec.Console = override.Console
	}
	for _, media := range append([]string{spec.Media.ISO}, spec.Media.ExtraISOs...) {
		info, e := os.Stat(media)
		if e != nil {
			return record, e
		}
		if !info.Mode().IsRegular() {
			return record, fmt.Errorf("%w: media is not a file", ErrState)
		}
	}
	for file, required := range map[string]bool{"nvram.fd": spec.Machine.Firmware == "uefi", "tpm": spec.Machine.TPM} {
		if required {
			if _, e := os.Stat(filepath.Join(tmp, file)); e != nil {
				return record, e
			}
		}
	}
	spec.Name = opts.Name
	record.Name = opts.Name
	data, err = json.Marshal(spec)
	if err != nil {
		return record, err
	}
	if err = state.Write(filepath.Join(tmp, "spec.json"), data); err != nil {
		return record, err
	}
	data, err = json.Marshal(record)
	if err != nil {
		return record, err
	}
	if err = state.Write(filepath.Join(tmp, "state.json"), data); err != nil {
		return record, err
	}
	if err = os.Remove(filepath.Join(tmp, "manifest.json")); err != nil {
		return record, err
	}
	if err = ctx.Err(); err != nil {
		return record, err
	}
	err = os.Rename(tmp, target)
	return record, err
}
