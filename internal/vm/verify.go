package vm

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"vmctl/internal/state"
)

// Verify checks the kickstart completion marker over authenticated SSH without changing the guest.
func (m Manager) Verify(ctx context.Context, name string) (err error) {
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
	if instance.Spec.Install.Mode != "kickstart" || (instance.Record.Phase != state.Installed && instance.Record.Phase != state.Verified) {
		return fmt.Errorf("%w: verification requires a completed kickstart installation", ErrState)
	}
	container, err := m.Docker.InspectInstance(ctx, instance.Directory)
	if err != nil {
		return err
	}
	if !container.State.Running {
		return fmt.Errorf("%w: verification requires a running guest", ErrState)
	}
	var host string
	var port int
	for _, forward := range instance.Spec.Network.Forwards {
		if forward.Guest == 22 {
			host, port = forward.Bind, forward.Host
			break
		}
	}
	if port == 0 {
		return fmt.Errorf("%w: verification requires an SSH forward", ErrState)
	}
	if ip := net.ParseIP(host); ip != nil && ip.IsUnspecified() {
		host = "127.0.0.1"
	}
	bounded, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	command := exec.CommandContext(bounded, "ssh", "-p", strconv.Itoa(port),
		"-i", filepath.Join(instance.Directory, "ssh-key"),
		"-o", "BatchMode=yes", "-o", "IdentitiesOnly=yes", "-o", "ConnectTimeout=10",
		"-o", "StrictHostKeyChecking=accept-new",
		"-o", "UserKnownHostsFile="+filepath.Join(instance.Directory, "known_hosts"),
		"root@"+host, "test -f /root/vmctl-install-complete && mountpoint -q /mnt/iso")
	command.WaitDelay = time.Second
	output, runErr := command.CombinedOutput()
	if bounded.Err() != nil {
		return bounded.Err()
	}
	if runErr != nil {
		return fmt.Errorf("verify guest: %w: %s", runErr, output)
	}
	instance.Record.Phase = state.Verified
	return m.Store.Save(instance.Record)
}
