package vm

import (
	"context"
	"errors"
	"net"
	"testing"

	"vmctl/internal/config"
	"vmctl/internal/state"
)

func TestPortsRejectOccupiedListener(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	address := listener.Addr().(*net.TCPAddr)
	spec := config.Spec{Console: config.Console{VNCBind: "127.0.0.1", VNCPort: address.Port}}
	if err := checkPorts(spec); err == nil {
		t.Fatal("occupied console port accepted")
	}
}

func TestInstallRejectsVerifiedInstanceBeforeRuntimeAccess(t *testing.T) {
	m := Manager{Store: state.Store{Root: t.TempDir()}}
	r := state.Record{Name: "demo", Phase: state.Verified}
	if err := m.Store.Save(r); err != nil {
		t.Fatal(err)
	}
	dir, err := m.Store.Path("demo")
	if err != nil {
		t.Fatal(err)
	}
	if err := state.Write(dir+"/spec.json", []byte(`{"name":"demo"}`)); err != nil {
		t.Fatal(err)
	}
	if err := m.Start(context.Background(), "demo", true); !errors.Is(err, ErrState) {
		t.Fatalf("expected installation state rejection, got %v", err)
	}
}

func TestStartHonorsCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := (Manager{}).Start(ctx, "demo", false)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v", err)
	}
}
