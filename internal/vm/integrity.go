package vm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"vmctl/internal/config"
)

func readCredentials(path string) (Credentials, error) {
	var c Credentials
	f, err := os.Open(path)
	if err != nil {
		return c, err
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, 4097))
	if err != nil {
		return c, err
	}
	if len(data) > 4096 {
		return c, ErrState
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if err = d.Decode(&c); err != nil {
		return c, err
	}
	var extra json.RawMessage
	if d.Decode(&extra) != io.EOF {
		return c, ErrState
	}
	if len(c.VNCPassword) != 8 {
		return c, fmt.Errorf("%w: VNC password must contain eight printable ASCII bytes", ErrState)
	}
	for _, b := range []byte(c.VNCPassword) {
		if b < 33 || b > 126 {
			return c, ErrState
		}
	}
	return c, nil
}

func validatePersistentFiles(dir string, machine config.Machine) error {
	for file, required := range map[string]bool{"nvram.fd": machine.Firmware == "uefi", "tpm": machine.TPM} {
		info, err := os.Lstat(filepath.Join(dir, file))
		if errors.Is(err, os.ErrNotExist) && !required {
			continue
		}
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return ErrState
		}
		if file == "nvram.fd" {
			if !info.Mode().IsRegular() || info.Size() == 0 {
				return fmt.Errorf("%w: invalid NVRAM", ErrState)
			}
		} else {
			if !info.IsDir() {
				return fmt.Errorf("%w: TPM state must be a directory", ErrState)
			}
			entries, err := os.ReadDir(filepath.Join(dir, file))
			if err != nil {
				return err
			}
			if len(entries) == 0 {
				return fmt.Errorf("%w: empty TPM state", ErrState)
			}
		}
	}
	return nil
}

type contextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r contextReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.reader.Read(p)
}
func copyContext(ctx context.Context, dst io.Writer, src io.Reader) (int64, error) {
	return io.CopyBuffer(dst, contextReader{ctx, src}, make([]byte, 128*1024))
}

func (m Manager) validateBundleDisk(ctx context.Context, dir string) error {
	base := []string{"run", "--rm", "--network=none", "--user", fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid()), "--mount", "type=bind,src=" + dir + ",dst=/instance,readonly", m.Image, "qemu-img"}
	out, err := m.Docker.Run(ctx, append(base, "info", "--output=json", "/instance/disk.qcow2")...)
	if err != nil {
		return err
	}
	var info struct {
		Format   string `json:"format"`
		Backing  string `json:"backing-filename"`
		DataFile string `json:"data-file"`
	}
	if err = json.Unmarshal(out, &info); err != nil {
		return err
	}
	if info.Format != "qcow2" || info.Backing != "" || info.DataFile != "" {
		return fmt.Errorf("%w: disk must be a self-contained qcow2", ErrState)
	}
	_, err = m.Docker.Run(ctx, append(base, "check", "-f", "qcow2", "/instance/disk.qcow2")...)
	return err
}
