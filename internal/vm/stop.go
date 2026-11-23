package vm

import (
	"context"
	"errors"
	"path/filepath"
	"time"

	"vmctl/internal/docker"
	"vmctl/internal/qmp"
)

func (m Manager) Stop(ctx context.Context, name string, force bool) (err error) {
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
	if err != nil || !exists {
		return err
	}
	container, err := m.Docker.InspectInstance(ctx, instance.Directory)
	if err != nil || !container.State.Running {
		return err
	}
	if force {
		_, err := m.Docker.Run(ctx, "kill", docker.ContainerName(instance.Directory))
		return err
	}
	if err := qmp.Execute(ctx, filepath.Join(instance.Directory, "qmp.sock"), qmp.Command{Name: "system_powerdown"}); err != nil {
		return err
	}
	waitCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-waitCtx.Done():
			return waitCtx.Err()
		case <-ticker.C:
			container, err := m.Docker.InspectInstance(waitCtx, instance.Directory)
			if err != nil || !container.State.Running {
				return err
			}
		}
	}
}
