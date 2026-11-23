package vm

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func extractBundle(ctx context.Context, archive, dir string) error {
	f, err := os.Open(archive)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	seen := map[string]bool{}
	var total int64
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		h, e := tr.Next()
		if e == io.EOF {
			break
		}
		if e != nil {
			return e
		}
		if h.Typeflag != tar.TypeReg || seen[h.Name] || (!bundleAllowed(h.Name) && h.Name != "manifest.json") || h.Size < 0 {
			return fmt.Errorf("%w: unsafe archive entry %q", ErrState, h.Name)
		}
		total += h.Size
		if total > 1<<40 || len(seen) >= 10000 {
			return fmt.Errorf("%w: archive exceeds import limits", ErrState)
		}
		seen[h.Name] = true
		dst := filepath.Join(dir, h.Name)
		if err := os.MkdirAll(filepath.Dir(dst), 0700); err != nil {
			return err
		}
		out, e := os.OpenFile(dst, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if e != nil {
			return e
		}
		_, copyErr := io.Copy(out, tr)
		closeErr := out.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
	}
	data, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		return err
	}
	var manifest bundleManifest
	if err = json.Unmarshal(data, &manifest); err != nil {
		return err
	}
	if manifest.Version != 1 || len(manifest.Files) != len(seen)-1 {
		return fmt.Errorf("%w: invalid manifest", ErrState)
	}
	for file, digest := range manifest.Files {
		if !bundleAllowed(file) || !seen[file] {
			return fmt.Errorf("%w: invalid manifest path", ErrState)
		}
		got, e := bundleDigest(ctx, filepath.Join(dir, file))
		if e != nil {
			return e
		}
		if got != digest {
			return fmt.Errorf("%w: checksum mismatch for %s", ErrState, file)
		}
	}
	for _, file := range []string{"disk.qcow2", "state.json", "spec.json", "credentials.json"} {
		if !seen[file] {
			return fmt.Errorf("%w: missing %s", ErrState, file)
		}
	}
	return nil
}
