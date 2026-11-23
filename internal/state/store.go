package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"syscall"
)

var ErrBusy = errors.New("instance is locked")
var ErrPath = errors.New("unsafe instance path")

type Phase string

const (
	Created    Phase = "created"
	Installing Phase = "installing"
	Installed  Phase = "installed-unverified"
	Verified   Phase = "verified"
)

type Record struct {
	Name  string `json:"name"`
	Phase Phase  `json:"phase"`
	MAC   string `json:"mac"`
}

type Store struct{ Root string }

func (s Store) Path(name string) (string, error) {
	if !regexp.MustCompile(`^[a-z][a-z0-9-]{0,47}$`).MatchString(name) {
		return "", fmt.Errorf("%w: %q", ErrPath, name)
	}
	root, err := filepath.Abs(s.Root)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(root, 0700); err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", err
	}
	if resolved != root {
		return "", fmt.Errorf("%w: symbolic root", ErrPath)
	}
	p := filepath.Join(root, name)
	info, err := os.Lstat(p)
	if err == nil && (!info.IsDir() || info.Mode()&os.ModeSymlink != 0) {
		return "", fmt.Errorf("%w: %s", ErrPath, p)
	}
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	return p, nil
}

func (s Store) Lock(name string) (*os.File, error) {
	p, err := s.Path(name)
	if err != nil {
		return nil, err
	}
	file, err := os.OpenFile(filepath.Join(filepath.Dir(p), "."+name+".lock"), os.O_CREATE|os.O_RDWR|syscall.O_NOFOLLOW, 0600)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		closeErr := file.Close()
		return nil, errors.Join(ErrBusy, err, closeErr)
	}
	return file, nil
}

func (s Store) Save(record Record) error {
	p, err := s.Path(record.Name)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(p, 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return err
	}
	return Write(filepath.Join(p, "state.json"), data)
}

func (s Store) Load(name string) (record Record, err error) {
	p, err := s.Path(name)
	if err != nil {
		return record, err
	}
	file, err := os.OpenFile(filepath.Join(p, "state.json"), os.O_RDONLY|syscall.O_NOFOLLOW, 0)
	if err != nil {
		return record, err
	}
	defer func() { err = errors.Join(err, file.Close()) }()
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&record); err != nil {
		return record, err
	}
	var trailing json.RawMessage
	if trailingErr := decoder.Decode(&trailing); trailingErr != io.EOF {
		return record, fmt.Errorf("%w: trailing state data", ErrPath)
	}
	if record.Name != name {
		return record, fmt.Errorf("%w: record name mismatch", ErrPath)
	}
	switch record.Phase {
	case Created, Installing, Installed, Verified:
		return record, nil
	default:
		return record, fmt.Errorf("%w: unknown installation phase %q", ErrPath, record.Phase)
	}
}

func Write(path string, data []byte) error {
	file, err := os.CreateTemp(filepath.Dir(path), ".pending-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	if _, err := file.Write(data); err != nil {
		return errors.Join(err, file.Close())
	}
	if err := file.Sync(); err != nil {
		return errors.Join(err, file.Close())
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(file.Name(), path)
}
