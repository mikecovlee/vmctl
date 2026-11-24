package vm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"vmctl/internal/qmp"
)

// Checkpoint uses QEMU's internal VM snapshots, including RAM and device state.
// Firmware and external TPM state cannot be captured atomically by savevm.
func (m Manager) Checkpoint(ctx context.Context, name, label string, restore bool) (err error) {
	return m.CheckpointWithTimeout(ctx, name, label, restore, 5*time.Minute)
}

func (m Manager) CheckpointWithTimeout(ctx context.Context, name, label string, restore bool, timeout time.Duration) (err error) {
	if timeout <= 0 {
		return fmt.Errorf("%w: checkpoint timeout must be positive", ErrState)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	if !snapshotName.MatchString(label) {
		return fmt.Errorf("%w: invalid checkpoint name", ErrState)
	}
	lock, err := m.Store.Lock(name)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, lock.Close()) }()
	i, err := m.Load(name)
	if err != nil {
		return err
	}
	if i.Spec.Machine.TPM || i.Spec.Machine.Firmware != "bios" {
		return fmt.Errorf("%w: memory checkpoints currently require BIOS without TPM; use a stopped snapshot for UEFI/TPM guests", ErrState)
	}
	c, err := m.Docker.InspectInstance(ctx, i.Directory)
	if err != nil {
		return err
	}
	if !c.State.Running {
		return fmt.Errorf("%w: checkpoint requires a running QEMU process", ErrState)
	}
	run, err := queryRunState(ctx, i.Directory)
	if err != nil {
		return err
	}
	if run.Status != "running" && run.Status != "paused" {
		return fmt.Errorf("%w: unsupported runtime state %s", ErrState, run.Status)
	}
	verb := "savevm"
	if restore {
		verb = "loadvm"
	}
	args, err := json.Marshal(struct {
		Command string `json:"command-line"`
	}{verb + " " + label})
	if err != nil {
		return err
	}
	result, err := qmp.Query(ctx, filepath.Join(i.Directory, "qmp.sock"), qmp.Command{Name: "human-monitor-command", Arguments: args, Timeout: timeout})
	if err != nil {
		if errors.Is(err, qmp.ErrOutcomeUnknown) {
			return fmt.Errorf("checkpoint outcome unknown; QEMU may still be processing; inspect status before retrying: %w", err)
		}
		return err
	}
	var output string
	if err := json.Unmarshal(result, &output); err != nil {
		return err
	}
	if strings.TrimSpace(output) != "" {
		return fmt.Errorf("%w: QEMU %s: %s", ErrState, verb, output)
	}
	return nil
}
