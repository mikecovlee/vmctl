package vm

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"vmctl/internal/config"
	"vmctl/internal/docker"
	"vmctl/internal/state"
)

var ErrExists = errors.New("instance already exists")
var ErrState = errors.New("invalid instance state")

type Manager struct {
	Store  state.Store
	Docker docker.Client
	Image  string
}

type Credentials struct {
	VNCPassword string `json:"vncPassword"`
}

func (m Manager) Create(ctx context.Context, spec config.Spec) (record state.Record, err error) {
	if err := ctx.Err(); err != nil {
		return record, err
	}
	lock, err := m.Store.Lock(spec.Name)
	if err != nil {
		return record, err
	}
	defer func() { err = errors.Join(err, lock.Close()) }()
	dir, err := m.Store.Path(spec.Name)
	if err != nil {
		return record, err
	}
	if err := os.Mkdir(dir, 0700); err != nil {
		if errors.Is(err, os.ErrExist) {
			return record, ErrExists
		}
		return record, err
	}
	defer func() {
		if err != nil {
			err = errors.Join(err, os.RemoveAll(dir))
		}
	}()
	media := append([]string{spec.Media.ISO}, spec.Media.ExtraISOs...)
	for _, path := range media {
		info, statErr := os.Stat(path)
		if statErr != nil {
			return record, statErr
		}
		if !info.Mode().IsRegular() {
			return record, fmt.Errorf("%w: media must be a regular file", ErrState)
		}
	}
	var random [10]byte
	if _, err := rand.Read(random[:]); err != nil {
		return record, err
	}
	record = state.Record{Name: spec.Name, Phase: state.Created, MAC: fmt.Sprintf("52:54:00:%02x:%02x:%02x", random[0], random[1], random[2])}
	credentials := Credentials{VNCPassword: hex.EncodeToString(random[3:7])}
	credentialData, err := json.Marshal(credentials)
	if err != nil {
		return record, err
	}
	if err := state.Write(filepath.Join(dir, "credentials.json"), credentialData); err != nil {
		return record, err
	}
	specData, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		return record, err
	}
	if err := state.Write(filepath.Join(dir, "spec.json"), specData); err != nil {
		return record, err
	}
	if err := m.prepareDisk(ctx, Instance{Directory: dir, Spec: spec, Record: record}); err != nil {
		return record, err
	}
	if err := m.Store.Save(record); err != nil {
		return record, err
	}
	return record, nil
}
