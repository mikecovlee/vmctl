package vm

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"vmctl/internal/config"
	"vmctl/internal/state"
)

type Instance struct {
	Directory string       `json:"directory"`
	Spec      config.Spec  `json:"spec"`
	Record    state.Record `json:"state"`
}

func (m Manager) Load(name string) (Instance, error) {
	dir, err := m.Store.Path(name)
	if err != nil {
		return Instance{}, err
	}
	record, err := m.Store.Load(name)
	if err != nil {
		return Instance{}, err
	}
	data, err := os.ReadFile(filepath.Join(dir, "spec.json"))
	if err != nil {
		return Instance{}, err
	}
	var spec config.Spec
	if err := json.Unmarshal(data, &spec); err != nil {
		return Instance{}, err
	}
	if spec.Name != name {
		return Instance{}, fmt.Errorf("%w: instance name mismatch", ErrState)
	}
	return Instance{Directory: dir, Spec: spec, Record: record}, nil
}
