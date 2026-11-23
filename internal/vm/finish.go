package vm

import (
	"context"
	"errors"
	"fmt"

	"vmctl/internal/state"
)

func (m Manager) FinishInstall(ctx context.Context, name string) (err error) {
	if err := ctx.Err(); err != nil {
		return err
	}
	lock, err := m.Store.Lock(name)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, lock.Close()) }()
	record, err := m.Store.Load(name)
	if err != nil {
		return err
	}
	if record.Phase != state.Installing {
		return fmt.Errorf("%w: installation has not started", ErrState)
	}
	record.Phase = state.Installed
	return m.Store.Save(record)
}
