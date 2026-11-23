package vm

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type bundleManifest struct {
	Version int               `json:"version"`
	Files   map[string]string `json:"files"`
}

func bundleAllowed(name string) bool {
	switch name {
	case "disk.qcow2", "nvram.fd", "state.json", "spec.json", "credentials.json", "ssh-key", "ssh-key.pub":
		return true
	}
	return strings.HasPrefix(name, "tpm/") && filepath.ToSlash(filepath.Clean(name)) == name && !strings.Contains(name, "\\") && name != "tpm/socket"
}

func bundleDigest(ctx context.Context, path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	buf := make([]byte, 1024*1024)
	for {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		n, e := f.Read(buf)
		if n > 0 {
			if _, err := h.Write(buf[:n]); err != nil {
				return "", err
			}
		}
		if e == io.EOF {
			break
		}
		if e != nil {
			return "", e
		}
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func writeManifest(ctx context.Context, dir string) error {
	m := bundleManifest{Version: 1, Files: map[string]string{}}
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if d.IsDir() {
			return nil
		}
		rel, e := filepath.Rel(dir, path)
		if e != nil {
			return e
		}
		if !bundleAllowed(rel) {
			return fmt.Errorf("%w: unexpected bundle file %s", ErrState, rel)
		}
		digest, e := bundleDigest(ctx, path)
		if e != nil {
			return e
		}
		m.Files[rel] = digest
		return nil
	})
	if err != nil {
		return err
	}
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "manifest.json"), data, 0600)
}
