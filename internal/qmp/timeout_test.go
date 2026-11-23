package qmp

import (
	"context"
	"net"
	"path/filepath"
	"testing"
	"time"
)

func TestSilentMonitorHonorsDeadline(t *testing.T) {
	path := filepath.Join(t.TempDir(), "qmp.sock")
	listener, err := net.Listen("unix", path)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	done := make(chan error, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			done <- err
			return
		}
		defer conn.Close()
		var data [1]byte
		_, err = conn.Read(data[:])
		done <- err
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	if err := Execute(ctx, path, Command{Name: "query-status"}); err == nil {
		t.Fatal("silent monitor reported success")
	}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("monitor connection remained open after deadline")
	}
}
