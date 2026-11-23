package vm

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestKVMIsolatedControlledEndpoints(t *testing.T) {
	if os.Getenv("VMCTL_KVM_TEST") != "1" {
		t.Skip("requires explicit isolated KVM fixture")
	}
	root, name := os.Getenv("VMCTL_TEST_STATE"), os.Getenv("VMCTL_TEST_INSTANCE")
	port := os.Getenv("VMCTL_TEST_SSH_PORT")
	expectReachable := os.Getenv("VMCTL_TEST_NETWORK") == "egress"
	if root == "" || name == "" || port == "" {
		t.Fatal("explicit fixture state, instance and SSH port required")
	}
	tcp, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer tcp.Close()
	udp, err := net.ListenPacket("udp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer udp.Close()
	go func() {
		for {
			c, err := tcp.Accept()
			if err != nil {
				return
			}
			_, _ = c.Write([]byte("vmctl-probe"))
			c.Close()
		}
	}()
	go func() {
		buf := make([]byte, 512)
		for {
			n, addr, err := udp.ReadFrom(buf)
			if err != nil {
				return
			}
			_, _ = udp.WriteTo(buf[:n], addr)
		}
	}()
	for _, endpoint := range []struct{ protocol, address string }{{"tcp4", tcp.Addr().String()}, {"udp4", udp.LocalAddr().String()}} {
		c, err := net.DialTimeout(endpoint.protocol, endpoint.address, time.Second)
		if err != nil {
			t.Fatal(err)
		}
		if err := c.SetDeadline(time.Now().Add(time.Second)); err != nil {
			t.Fatal(err)
		}
		if endpoint.protocol == "udp4" {
			if _, err := c.Write([]byte("vmctl-probe")); err != nil {
				t.Fatal(err)
			}
		}
		buf := make([]byte, 32)
		n, err := c.Read(buf)
		c.Close()
		if err != nil || string(buf[:n]) != "vmctl-probe" {
			t.Fatalf("control endpoint failed: %q %v", buf[:n], err)
		}
	}
	script := fmt.Sprintf(`import socket
import subprocess
routes=subprocess.run(['ip','-6','route','show','default'],check=True,capture_output=True,text=True)
if routes.stdout.strip(): raise SystemExit('unexpected IPv6 default route: '+routes.stdout)
expected=%s
for kind, port in [(socket.SOCK_STREAM,%d),(socket.SOCK_DGRAM,%d)]:
    s=socket.socket(socket.AF_INET,kind)
    s.settimeout(2)
    try:
        s.connect(('10.0.2.2',port))
        if kind == socket.SOCK_DGRAM: s.send(b'vmctl-probe')
        data=s.recv(32)
    except (socket.timeout, ConnectionRefusedError):
        if expected: raise SystemExit('control endpoint unreachable in egress mode')
    else:
        if not expected: raise SystemExit('unexpected reachable endpoint: '+repr(data))
        if data != b'vmctl-probe': raise SystemExit('unexpected endpoint response')
    finally:
        s.close()
print('CONTROLLED_ISOLATION_OK')
`, map[bool]string{true: "True", false: "False"}[expectReachable], tcp.Addr().(*net.TCPAddr).Port, udp.LocalAddr().(*net.UDPAddr).Port)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	dir := filepath.Join(root, name)
	cmd := exec.CommandContext(ctx, "ssh", "-p", port, "-i", filepath.Join(dir, "ssh-key"), "-o", "BatchMode=yes", "-o", "StrictHostKeyChecking=yes", "-o", "UserKnownHostsFile="+filepath.Join(dir, "known_hosts"), "root@127.0.0.1", "python3 -")
	cmd.Stdin = strings.NewReader(script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("guest probe failed: %v: %s", err, output)
	}
	if !strings.Contains(string(output), "CONTROLLED_ISOLATION_OK") {
		t.Fatalf("missing result: %s", output)
	}
}
