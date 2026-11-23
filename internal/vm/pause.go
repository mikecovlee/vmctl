package vm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"vmctl/internal/qmp"
)

type RunState struct {
	Running bool   `json:"running"`
	Status  string `json:"status"`
}

func queryRunState(ctx context.Context, directory string) (RunState, error) {
	data, err := qmp.Query(ctx, filepath.Join(directory, "qmp.sock"), qmp.Command{Name: "query-status"})
	if err != nil {
		return RunState{}, err
	}
	var result RunState
	if err := json.Unmarshal(data, &result); err != nil {
		return result, err
	}
	if result.Status == "" {
		return result, fmt.Errorf("%w: missing runtime status", ErrState)
	}
	return result, nil
}

func (m Manager) Pause(ctx context.Context, name string, resume bool) (err error) {
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
	container, err := m.Docker.InspectInstance(ctx, instance.Directory)
	if err != nil {
		return err
	}
	if !container.State.Running {
		return fmt.Errorf("%w: instance is stopped", ErrState)
	}
	current, err := queryRunState(ctx, instance.Directory)
	if err != nil {
		return err
	}
	command := "stop"
	target := "paused"
	if resume {
		command, target = "cont", "running"
	}
	if current.Status == target {
		return nil
	}
	if current.Status != "running" && current.Status != "paused" {
		return fmt.Errorf("%w: cannot change runtime state %s", ErrState, current.Status)
	}
	return qmp.Execute(ctx, filepath.Join(instance.Directory, "qmp.sock"), qmp.Command{Name: command})
}
