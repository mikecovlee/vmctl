package docker

import (
	"strings"
	"testing"
	"vmctl/internal/config"
)

func TestQEMUDefaultsBlockOutbound(t *testing.T) {
	s, err := config.Parse(strings.NewReader(`{"apiVersion":"v1alpha1","name":"demo","resources":{"cpu":2,"memoryMiB":2048,"diskGiB":20},"media":{"iso":"installer.iso"}}`))
	if err != nil {
		t.Fatal(err)
	}
	args := QEMUArgs(s, "52:54:00:12:34:56", false)
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "restrict=on") || strings.Contains(joined, "guestfwd") {
		t.Fatalf("unsafe network: %s", joined)
	}
	if !strings.Contains(joined, "127.0.0.1:1,password=on") {
		t.Fatalf("unexpected console: %s", joined)
	}
}

func TestManualInstallBootsMediaOnce(t *testing.T) {
	s := config.Spec{Machine: config.Machine{DiskBus: "sata", NIC: "e1000"}, Install: config.Install{Mode: "manual"}, Network: config.Network{Mode: "egress"}, Console: config.Console{VNCBind: "127.0.0.1", VNCPort: 5902}}
	args := strings.Join(QEMUArgs(s, "52:54:00:12:34:57", true), " ")
	if !strings.Contains(args, "order=c,once=d") || !strings.Contains(args, "if=ide") || strings.Contains(args, "restrict=on") {
		t.Fatalf("unexpected boot: %s", args)
	}
}
