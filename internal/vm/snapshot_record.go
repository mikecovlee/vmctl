package vm

import (
	"encoding/json"
	"errors"
	"io"
	"os"

	"vmctl/internal/state"
)

func validateSnapshotRecord(path, name string) (err error) {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, f.Close()) }()
	d := json.NewDecoder(f)
	d.DisallowUnknownFields()
	var r state.Record
	if err := d.Decode(&r); err != nil {
		return err
	}
	if r.Name != name {
		return ErrState
	}
	switch r.Phase {
	case state.Created, state.Installing, state.Installed, state.Verified:
	default:
		return ErrState
	}
	var extra json.RawMessage
	if err := d.Decode(&extra); !errors.Is(err, io.EOF) {
		return ErrState
	}
	return nil
}
