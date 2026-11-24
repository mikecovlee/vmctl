package vm

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"vmctl/internal/config"
	"vmctl/internal/docker"
	"vmctl/internal/state"
)

func TestNetworkChangesPreserveIdentityAndRejectInvalidMode(t *testing.T) {
	root := t.TempDir()
	iso := filepath.Join(root, "input.iso")
	if err := os.WriteFile(iso, []byte("iso"), 0600); err != nil {
		t.Fatal(err)
	}
	spec, err := config.Parse(strings.NewReader(`{"apiVersion":"v1alpha1","name":"demo","resources":{"cpu":1,"memoryMiB":512,"diskGiB":2},"media":{"iso":"` + iso + `"}}`))
	if err != nil {
		t.Fatal(err)
	}
	m := Manager{Store: state.Store{Root: filepath.Join(root, "state")}, Docker: docker.Client{Binary: "/bin/true"}}
	before, err := m.Create(context.Background(), spec)
	if err != nil {
		t.Fatal(err)
	}
	if err := m.ChangeNetwork(context.Background(), "demo", NetworkChange{Mode: "nat"}); err != nil {
		t.Fatal(err)
	}
	if err := m.ChangeNetwork(context.Background(), "demo", NetworkChange{Mode: "invalid"}); err == nil {
		t.Fatal("invalid mode accepted")
	}
	i, err := m.Load("demo")
	if err != nil {
		t.Fatal(err)
	}
	if i.Spec.Network.Mode != "nat" || i.Record.MAC != before.MAC {
		t.Fatal("network or identity changed unexpectedly")
	}
	if err := m.ChangeNetwork(context.Background(), "demo", NetworkChange{Mode: "isolated"}); err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	port := listener.Addr().(*net.TCPAddr).Port
	if err := m.ChangeNetwork(context.Background(), "demo", NetworkChange{Add: &config.Forward{Bind: "127.0.0.1", Host: port, Guest: 8080}}); err == nil {
		t.Fatal("occupied port accepted")
	}
	if err := m.ChangeNetwork(context.Background(), "demo", NetworkChange{Add: &config.Forward{Bind: "127.0.0.1", Host: 0, Guest: 8080}}); err == nil {
		t.Fatal("invalid port accepted")
	}
	if err := m.ChangeNetwork(context.Background(), "demo", NetworkChange{Remove: &config.Forward{Bind: "127.0.0.1", Host: 12345}}); err == nil {
		t.Fatal("missing rule removal accepted")
	}
	n, err := m.Network(context.Background(), "demo")
	if err != nil {
		t.Fatal(err)
	}
	if len(n.Forwards) != 0 || n.Mode != "isolated" {
		t.Fatalf("failed mutation changed configuration: %+v", n)
	}
}
