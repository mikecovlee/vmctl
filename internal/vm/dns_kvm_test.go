package vm

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestKVMControlledDNS(t *testing.T) {
	if os.Getenv("VMCTL_KVM_TEST") != "1" {
		t.Skip("requires explicit KVM fixture")
	}
	root, name, port := os.Getenv("VMCTL_TEST_STATE"), os.Getenv("VMCTL_TEST_INSTANCE"), os.Getenv("VMCTL_TEST_SSH_PORT")
	if root == "" || name == "" || port == "" {
		t.Fatal("explicit fixture required")
	}
	query, err := hex.DecodeString("12340100000100000000000005766d63746c04746573740000010001")
	if err != nil {
		t.Fatal(err)
	}
	response := append([]byte(nil), query...)
	response[2], response[3] = 0x81, 0x83 // NXDOMAIN for the reserved .test name.
	server, err := net.ListenPacket("udp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		buf := make([]byte, 512)
		for {
			n, addr, err := server.ReadFrom(buf)
			if err != nil {
				return
			}
			if !bytes.Equal(buf[:n], query) {
				continue
			}
			if _, err := server.WriteTo(response, addr); err != nil {
				return
			}
		}
	}()
	t.Cleanup(func() {
		if err := server.Close(); err != nil {
			t.Error(err)
		}
		<-done
	})
	control, err := net.DialTimeout("udp4", server.LocalAddr().String(), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer control.Close()
	if err := control.SetDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	if _, err := control.Write(query); err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, 512)
	n, err := control.Read(buf)
	if err != nil || !bytes.Equal(buf[:n], response) {
		t.Fatalf("DNS control failed: %v", err)
	}
	expected := "False"
	if os.Getenv("VMCTL_TEST_NETWORK") == "egress" {
		expected = "True"
	}
	script := fmt.Sprintf(`import socket
s=socket.socket(socket.AF_INET,socket.SOCK_DGRAM)
s.settimeout(2)
try:
    s.sendto(bytes.fromhex('%s'),('10.0.2.2',%d))
    data,addr=s.recvfrom(512)
except (socket.timeout,ConnectionRefusedError):
    if %s: raise SystemExit('DNS endpoint unreachable in egress mode')
else:
    if not %s: raise SystemExit('DNS leaked in isolated mode')
    if data != bytes.fromhex('%s'): raise SystemExit('invalid DNS response')
finally:
    s.close()
print('CONTROLLED_DNS_OK')
`, hex.EncodeToString(query), server.LocalAddr().(*net.UDPAddr).Port, expected, expected, hex.EncodeToString(response))
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	dir := filepath.Join(root, name)
	cmd := exec.CommandContext(ctx, "ssh", "-p", port, "-i", filepath.Join(dir, "ssh-key"), "-o", "BatchMode=yes", "-o", "StrictHostKeyChecking=yes", "-o", "UserKnownHostsFile="+filepath.Join(dir, "known_hosts"), "root@127.0.0.1", "python3 -")
	cmd.Stdin = strings.NewReader(script)
	output, err := cmd.CombinedOutput()
	if err != nil || !strings.Contains(string(output), "CONTROLLED_DNS_OK") {
		t.Fatalf("DNS probe: %v: %s", err, output)
	}
}
