package qmp

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"path/filepath"
	"testing"
	"time"
)

func TestCommandRejectsProtocolFailures(t *testing.T) {
	for _, reply := range []string{`{"id":"wrong","return":{}}`, `{"id":"vmctl-0","error":{"class":"CommandNotFound","desc":"unsupported"}}`} {
		t.Run(reply, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "monitor.sock")
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
				enc := json.NewEncoder(conn)
				if err := enc.Encode(json.RawMessage(`{"QMP":{}}`)); err != nil {
					done <- err
					return
				}
				var request json.RawMessage
				if err := json.NewDecoder(conn).Decode(&request); err != nil {
					done <- err
					return
				}
				done <- enc.Encode(json.RawMessage(reply))
			}()
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			if err := Execute(ctx, path, Command{Name: "query-status"}); !errors.Is(err, ErrProtocol) {
				t.Fatalf("expected protocol failure, got %v", err)
			}
			if err := <-done; err != nil {
				t.Fatal(err)
			}
		})
	}
}
