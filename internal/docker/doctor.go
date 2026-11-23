package docker

import (
	"context"
	"fmt"
	"runtime"
)

// Doctor checks acceleration inside the same Docker device boundary as instances.
func (c Client) Doctor(ctx context.Context, image string) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("%w: Linux host required", ErrCommand)
	}
	if _, err := c.Run(ctx, "info", "--format", "{{.ServerVersion}}"); err != nil {
		return err
	}
	_, err := c.Run(ctx, "run", "--rm", "--network=none", "--device=/dev/kvm", image,
		"sh", "-ec", `test -r /usr/share/OVMF/OVMF_CODE_4M.fd
test -r /usr/share/OVMF/OVMF_VARS_4M.ms.fd
command -v swtpm >/dev/null
qemu-system-x86_64 -machine none,accel=kvm -display none -nodefaults -S -daemonize -pidfile /tmp/kvm.pid
kill "$(cat /tmp/kvm.pid)"`)
	return err
}
