package vm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"

	"vmctl/internal/config"
	"vmctl/internal/state"
)

type NetworkChange struct {
	Mode   string
	Add    *config.Forward
	Remove *config.Forward
}

func (m Manager) Network(ctx context.Context, name string) (config.Network, error) {
	if err := ctx.Err(); err != nil {
		return config.Network{}, err
	}
	i, err := m.Load(name)
	if err != nil {
		return config.Network{}, err
	}
	n := i.Spec.Network
	if n.Mode == "egress" {
		n.Mode = "nat"
	}
	return n, nil
}

func (m Manager) ChangeNetwork(ctx context.Context, name string, change NetworkChange) (err error) {
	if err = ctx.Err(); err != nil {
		return err
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
	exists, err := m.Docker.Exists(ctx, i.Directory)
	if err != nil {
		return err
	}
	if exists {
		c, e := m.Docker.InspectInstance(ctx, i.Directory)
		if e != nil {
			return e
		}
		if c.State.Running {
			return fmt.Errorf("%w: stop instance before changing network", ErrState)
		}
	}
	if change.Mode != "" {
		i.Spec.Network.Mode = change.Mode
	}
	if change.Add != nil {
		i.Spec.Network.Forwards = append(i.Spec.Network.Forwards, *change.Add)
	}
	if change.Remove != nil {
		found := false
		forwards := make([]config.Forward, 0, len(i.Spec.Network.Forwards))
		for _, f := range i.Spec.Network.Forwards {
			if f.Bind == change.Remove.Bind && f.Host == change.Remove.Host {
				found = true
				continue
			}
			forwards = append(forwards, f)
		}
		if !found {
			return fmt.Errorf("%w: forwarding rule not found", ErrState)
		}
		i.Spec.Network.Forwards = forwards
	}
	data, err := json.Marshal(i.Spec)
	if err != nil {
		return err
	}
	spec, err := config.Parse(bytes.NewReader(data))
	if err != nil {
		return err
	}
	if change.Add != nil {
		if err := checkPorts(spec); err != nil {
			return err
		}
	}
	data, err = json.Marshal(spec)
	if err != nil {
		return err
	}
	return state.Write(filepath.Join(i.Directory, "spec.json"), data)
}
