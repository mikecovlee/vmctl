package docker

import (
	"crypto/sha256"
	"fmt"
	"os"
	"syscall"

	"vmctl/internal/config"
)

func ContainerName(instance string) string {
	sum := sha256.Sum256([]byte(instance))
	return fmt.Sprintf("vmctl-%x", sum[:12])
}

func RunArgs(spec config.Spec, instance, image string) []string {
	args := []string{"run", "--detach", "--interactive", "--name", ContainerName(instance),
		"--user", fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid()),
		"--restart=no", "--network=host", "--device=/dev/kvm",
		"--label", "io.vmctl.managed=true", "--label", "io.vmctl.instance=" + instance,
		"--mount", "type=bind,src=" + instance + ",dst=/instance",
		"--mount", "type=bind,src=" + spec.Media.ISO + ",dst=/media/install.iso,readonly"}
	if info, err := os.Stat("/dev/kvm"); err == nil {
		if stat, ok := info.Sys().(*syscall.Stat_t); ok {
			args = append(args, "--group-add", fmt.Sprint(stat.Gid))
		}
	}
	for i, path := range spec.Media.ExtraISOs {
		args = append(args, "--mount", fmt.Sprintf("type=bind,src=%s,dst=/media/extra-%d.iso,readonly", path, i))
	}
	if spec.Machine.TPM {
		args = append(args, "--env", "VMCTL_TPM=true")
	}
	return append(args, image)
}
