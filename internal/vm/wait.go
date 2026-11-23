package vm

import (
	"context"
	"errors"
	"fmt"
	"time"

	"vmctl/internal/state"
)

func (m Manager) WaitInstall(ctx context.Context, name string) (err error) {
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
	if instance.Spec.Install.Mode != "kickstart" || instance.Record.Phase != state.Installing {
		return fmt.Errorf("%w: wait requires an active kickstart installation", ErrState)
	}
	bounded, cancel := context.WithTimeout(ctx, 45*time.Minute)
	defer cancel()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		container, err := m.Docker.InspectInstance(bounded, instance.Directory)
		if err != nil {
			return err
		}
		if !container.State.Running {
			if container.State.ExitCode != 0 {
				return fmt.Errorf("%w: installer exited %d", ErrState, container.State.ExitCode)
			}
			instance.Record.Phase = state.Installed
			return m.Store.Save(instance.Record)
		}
		select {
		case <-bounded.Done():
			return bounded.Err()
		case <-ticker.C:
		}
	}
}
