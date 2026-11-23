package vm

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"vmctl/internal/install"
	"vmctl/internal/state"
)

func (m Manager) prepareKickstart(ctx context.Context, i Instance) error {
	boot := filepath.Join(i.Directory, "boot")
	if err := os.MkdirAll(boot, 0700); err != nil {
		return err
	}
	base := []string{"run", "--rm", "--user", fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid()), "--mount", "type=bind,src=" + i.Directory + ",dst=/instance", "--mount", "type=bind,src=" + i.Spec.Media.ISO + ",dst=/media/install.iso,readonly", m.Image}
	key := filepath.Join(i.Directory, "ssh-key")
	if _, err := os.Stat(key); os.IsNotExist(err) {
		if output, err := exec.CommandContext(ctx, "ssh-keygen", "-t", "ed25519", "-N", "", "-f", key).CombinedOutput(); err != nil {
			return fmt.Errorf("generate instance SSH key: %w: %s", err, output)
		}
	} else if err != nil {
		return err
	}
	public, err := os.ReadFile(key + ".pub")
	if err != nil {
		return err
	}
	source, err := os.ReadFile(i.Spec.Install.Template)
	if err != nil {
		return err
	}
	rendered, err := install.Render(string(source), install.Identity{Hostname: i.Spec.Name, MAC: i.Record.MAC, PublicKey: strings.TrimSpace(string(public))})
	if err != nil {
		return err
	}
	if err := state.Write(filepath.Join(boot, "ks.cfg"), []byte(rendered)); err != nil {
		return err
	}
	// Paths are positional arguments rather than interpolated shell source.
	script := `set -eu
7z e -so /media/install.iso "$1" > /instance/boot/vmlinuz
7z e -so /media/install.iso "$2" > /instance/boot/initrd.img
cd /instance/boot
printf 'ks.cfg\n' | cpio -H newc -o > ks.cpio
cat initrd.img ks.cpio > initrd-ks.img
`
	if _, err := m.Docker.Run(ctx, append(base, "sh", "-c", script, "prepare", i.Spec.Install.KernelPath, i.Spec.Install.InitrdPath)...); err != nil {
		return err
	}
	media, err := os.Open(i.Spec.Media.ISO)
	if err != nil {
		return err
	}
	digest, hashErr := install.Digest(media, rendered)
	if err := errors.Join(hashErr, media.Close()); err != nil {
		return err
	}
	return state.Write(filepath.Join(boot, "installation.sha256"), []byte(digest+"\n"))
}
