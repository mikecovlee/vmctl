package vm

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"vmctl/internal/docker"
	"vmctl/internal/state"
)

var ErrConfirmation = errors.New("destructive operation requires --yes")

type Status struct {
	Name        string      `json:"name"`
	Phase       state.Phase `json:"installation"`
	Running     bool        `json:"running"`
	Runtime     string      `json:"runtime"`
	NetworkMode string      `json:"networkMode"`
}

func (m Manager) Status(ctx context.Context, name string) (Status, error) {
	if err := ctx.Err(); err != nil {
		return Status{}, err
	}
	instance, err := m.Load(name)
	if err != nil {
		return Status{}, err
	}
	result := Status{Name: name, Phase: instance.Record.Phase, Runtime: "stopped"}
	result.NetworkMode = instance.Spec.Network.Mode
	if result.NetworkMode == "egress" {
		result.NetworkMode = "nat"
	}
	exists, err := m.Docker.Exists(ctx, instance.Directory)
	if err != nil || !exists {
		return result, err
	}
	container, err := m.Docker.InspectInstance(ctx, instance.Directory)
	if err != nil {
		return Status{}, err
	}
	result.Running = container.State.Running
	if container.State.Running {
		runtime, err := queryRunState(ctx, instance.Directory)
		if err != nil {
			return Status{}, err
		}
		result.Runtime = runtime.Status
	}
	return result, nil
}

func (m Manager) List(ctx context.Context) ([]Status, error) {
	entries, err := os.ReadDir(m.Store.Root)
	if errors.Is(err, os.ErrNotExist) {
		return []Status{}, nil
	}
	if err != nil {
		return nil, err
	}
	results := make([]Status, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		value, err := m.Status(ctx, entry.Name())
		if err != nil {
			return nil, err
		}
		results = append(results, value)
	}
	return results, nil
}

func (m Manager) Delete(ctx context.Context, name string, confirmed bool) (err error) {
	if !confirmed {
		return ErrConfirmation
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	lock, err := m.Store.Lock(name)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, lock.Close()) }()
	instance, err := m.Load(name)
	if err != nil {
		return err
	}
	exists, err := m.Docker.Exists(ctx, instance.Directory)
	if err != nil {
		return err
	}
	if exists {
		container, err := m.Docker.InspectInstance(ctx, instance.Directory)
		if err != nil {
			return err
		}
		if container.State.Running {
			return ErrState
		}
		if _, err := m.Docker.Run(ctx, "rm", docker.ContainerName(instance.Directory)); err != nil {
			return err
		}
	}
	return os.RemoveAll(filepath.Clean(instance.Directory))
}
