package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestReleaseRejectsInvalidInputsBeforeBuild(t *testing.T) {
	for _, input := range []struct{ version, commit string }{
		{"../escape", "0123456789012345678901234567890123456789"},
		{"v1.0.0", "not-a-commit"},
		{"v1.0.0-rc.1 extra", "0123456789012345678901234567890123456789"},
	} {
		t.Run(input.version+input.commit, func(t *testing.T) {
			root := t.TempDir()
			source, err := os.ReadFile("release.sh")
			if err != nil {
				t.Fatal(err)
			}
			if err := os.Mkdir(filepath.Join(root, "scripts"), 0700); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(root, "scripts/release.sh")
			if err := os.WriteFile(path, source, 0600); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command("sh", path, input.version)
			cmd.Env = append(os.Environ(), "VMCTL_COMMIT="+input.commit)
			if output, err := cmd.CombinedOutput(); err == nil {
				t.Fatalf("invalid release accepted: %s", output)
			}
			if _, err := os.Stat(filepath.Join(root, "dist")); !os.IsNotExist(err) {
				t.Fatalf("invalid input created release output: %v", err)
			}
		})
	}
}
