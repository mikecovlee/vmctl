package docker

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"vmctl/internal/config"
)

func TestRunArgsKeepsHostOwnership(t *testing.T) {
	args := strings.Join(RunArgs(config.Spec{}, "/private/demo", "runtime:test"), "\n")
	want := fmt.Sprintf("--user\n%d:%d", os.Getuid(), os.Getgid())
	if !strings.Contains(args, want) {
		t.Fatalf("host ownership missing: %s", args)
	}
}

func TestRunArgsIsolatesInstanceMounts(t *testing.T) {
	// Given an instance with installer media outside its state directory.
	spec := config.Spec{Name: "demo", Media: config.Media{ISO: "/media source/linux.iso"}}
	// When building the container invocation.
	args := RunArgs(spec, "/private/demo", "runtime:test")
	// Then only explicit mounts are exposed and restart is disabled.
	joined := strings.Join(args, "\n")
	for _, want := range []string{"--restart=no", "io.vmctl.instance=/private/demo", "type=bind,src=/media source/linux.iso,dst=/media/install.iso,readonly", "type=bind,src=/private/demo,dst=/instance"} {
		if !strings.Contains(joined, want) {
			t.Errorf("missing %q in %v", want, args)
		}
	}
}

func TestRunArgsEnablesRequestedTPM(t *testing.T) {
	spec := config.Spec{Machine: config.Machine{TPM: true}}
	args := strings.Join(RunArgs(spec, "/private/demo", "runtime:test"), "\n")
	if !strings.Contains(args, "VMCTL_TPM=true") {
		t.Fatal("requested TPM was not passed to runtime")
	}
}
