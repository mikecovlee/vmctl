package docker

import (
	"context"
	"testing"
)

func TestDoctorRejectsUnavailableRuntime(t *testing.T) {
	client := Client{Binary: "/bin/false"}
	if err := client.Doctor(context.Background(), "missing-runtime"); err == nil {
		t.Fatal("unavailable Docker must not report usable KVM")
	}
}
