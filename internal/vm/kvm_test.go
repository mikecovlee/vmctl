package vm

import (
	"context"
	"os"
	"testing"
	"time"

	"vmctl/internal/docker"
	"vmctl/internal/state"
)

func TestKVMGracefulLifecycle(t *testing.T) {
	if os.Getenv("VMCTL_KVM_TEST") != "1" {
		t.Skip("requires a dedicated installed Kickstart instance")
	}
	root, name := os.Getenv("VMCTL_TEST_STATE"), os.Getenv("VMCTL_TEST_INSTANCE")
	if root == "" || name == "" {
		t.Fatal("VMCTL_TEST_STATE and VMCTL_TEST_INSTANCE are required")
	}
	m := Manager{Store: state.Store{Root: root}, Docker: docker.Client{}, Image: "vmctl-runtime:local"}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	// Given an explicitly selected, independently installed test guest.
	before, err := m.Status(ctx, name)
	if err != nil || !before.Running || before.Phase != state.Verified {
		t.Fatalf("requires verified running guest: %+v, %v", before, err)
	}
	// When requesting ACPI shutdown without force.
	if err := m.Stop(ctx, name, false); err != nil {
		t.Fatal(err)
	}
	// Then the guest stops while retaining its installation identity.
	after, err := m.Status(ctx, name)
	if err != nil || after.Running || after.Phase != state.Verified {
		t.Fatalf("unexpected stopped state: %+v, %v", after, err)
	}
}
