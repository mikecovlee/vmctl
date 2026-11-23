package vm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"vmctl/internal/docker"
	"vmctl/internal/qmp"
	"vmctl/internal/state"
)

func (m Manager) Start(ctx context.Context, name string, installing bool) (err error) {
	if err := ctx.Err(); err != nil {
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
	if installing && i.Record.Phase != state.Created {
		return fmt.Errorf("%w: installation already started", ErrState)
	}
	if !installing && i.Record.Phase == state.Created {
		return fmt.Errorf("%w: install the instance first", ErrState)
	}
	credentials, err := readCredentials(filepath.Join(i.Directory, "credentials.json"))
	if err != nil {
		return err
	}
	exists, err := m.Docker.Exists(ctx, i.Directory)
	if err != nil {
		return err
	}
	if exists {
		c, err := m.Docker.InspectInstance(ctx, i.Directory)
		if err != nil {
			return err
		}
		if c.State.Running {
			if installing {
				return m.finishLaunch(ctx, i, credentials, true)
			}
			return nil
		}
	}
	if err := checkPorts(i.Spec); err != nil {
		return err
	}
	if installing {
		if i.Spec.Install.Mode == "kickstart" {
			if err := m.prepareKickstart(ctx, i); err != nil {
				return err
			}
		}
	}
	if exists {
		if _, err := m.Docker.Run(ctx, "rm", docker.ContainerName(i.Directory)); err != nil {
			return err
		}
	}
	args := docker.RunArgs(i.Spec, i.Directory, m.Image)
	args = append(args, "qemu-system-x86_64")
	args = append(args, docker.QEMUArgs(i.Spec, i.Record.MAC, installing)...)
	if _, err := m.Docker.Run(ctx, args...); err != nil {
		return err
	}
	return m.finishLaunch(ctx, i, credentials, installing)
}

func (m Manager) finishLaunch(ctx context.Context, i Instance, credentials Credentials, installing bool) error {
	arguments, err := json.Marshal(struct {
		Protocol string `json:"protocol"`
		Password string `json:"password"`
	}{"vnc", credentials.VNCPassword})
	if err != nil {
		return err
	}
	ready, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	var last error
	for {
		last = qmp.Execute(ready, filepath.Join(i.Directory, "qmp.sock"), qmp.Command{Name: "set_password", Arguments: arguments})
		if last == nil {
			if installing {
				i.Record.Phase = state.Installing
				return m.Store.Save(i.Record)
			}
			return nil
		}
		select {
		case <-ready.Done():
			return errors.Join(ready.Err(), last)
		case <-ticker.C:
		}
	}
}

func (m Manager) prepareDisk(ctx context.Context, i Instance) error {
	disk := filepath.Join(i.Directory, "disk.qcow2")
	if _, err := os.Lstat(disk); err == nil {
		return fmt.Errorf("%w: disk already exists; refusing overwrite", ErrExists)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	args := []string{"run", "--rm", "--user", fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid()), "--mount", "type=bind,src=" + i.Directory + ",dst=/instance", m.Image,
		"qemu-img", "create", "-f", "qcow2", "/instance/disk.qcow2", fmt.Sprintf("%dG", i.Spec.Resources.DiskGiB)}
	if _, err := m.Docker.Run(ctx, args...); err != nil {
		return err
	}
	if i.Spec.Machine.Firmware == "uefi" {
		vars := "/usr/share/OVMF/OVMF_VARS_4M.fd"
		if i.Spec.Machine.SecureBoot {
			vars = "/usr/share/OVMF/OVMF_VARS_4M.ms.fd"
		}
		_, err := m.Docker.Run(ctx, "run", "--rm", "--user", fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid()), "--mount", "type=bind,src="+i.Directory+",dst=/instance", m.Image, "cp", vars, "/instance/nvram.fd")
		return err
	}
	return nil
}
