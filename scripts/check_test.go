package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestShellSyntaxFailureStopsRepositoryChecks(t *testing.T) {
	script, err := os.ReadFile("check.sh")
	if err != nil {
		t.Fatal(err)
	}
	for _, invalid := range []bool{false, true} {
		t.Run(map[bool]string{false: "valid", true: "invalid"}[invalid], func(t *testing.T) {
			root := t.TempDir()
			for _, dir := range []string{"scripts", "runtime", "examples", "bin"} {
				if err := os.Mkdir(filepath.Join(root, dir), 0700); err != nil {
					t.Fatal(err)
				}
			}
			source := "#!/bin/sh\nexit 0\n"
			if invalid {
				source = "#!/bin/sh\nif then\n"
			}
			for path, contents := range map[string][]byte{
				"scripts/check.sh":    script,
				"examples/prepare.sh": []byte(source),
				"bin/go":              []byte("#!/bin/sh\nexit 77\n"),
			} {
				if err := os.WriteFile(filepath.Join(root, path), contents, 0700); err != nil {
					t.Fatal(err)
				}
			}
			cmd := exec.Command("sh", filepath.Join(root, "scripts/check.sh"))
			cmd.Env = append(os.Environ(), "PATH="+filepath.Join(root, "bin")+":"+os.Getenv("PATH"))
			out, err := cmd.CombinedOutput()
			var exit *exec.ExitError
			if !errors.As(err, &exit) {
				t.Fatalf("expected failure, got %v: %s", err, out)
			}
			if invalid && exit.ExitCode() == 77 {
				t.Fatalf("syntax error reached later checks: %s", out)
			}
			if !invalid && exit.ExitCode() != 77 {
				t.Fatalf("valid scripts failed syntax check: %s", out)
			}
		})
	}
}
