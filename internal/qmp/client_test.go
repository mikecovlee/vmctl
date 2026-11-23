package qmp

import (
	"context"
	"encoding/json"
	"net"
	"path/filepath"
	"testing"
	"time"
)

func TestCommandIgnoresEvents(t *testing.T) {
	// Given a monitor that emits an event before the command reply.
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
		encoder := json.NewEncoder(conn)
		decoder := json.NewDecoder(conn)
		if err := encoder.Encode(json.RawMessage(`{"QMP":{"version":{}}}`)); err != nil {
			done <- err
			return
		}
		for range 2 {
			var req struct {
				ID string `json:"id"`
			}
			if err := decoder.Decode(&req); err != nil {
				done <- err
				return
			}
			if err := encoder.Encode(json.RawMessage(`{"event":"STOP"}`)); err != nil {
				done <- err
				return
			}
			if err := encoder.Encode(struct {
				ID     string   `json:"id"`
				Return struct{} `json:"return"`
			}{ID: req.ID}); err != nil {
				done <- err
				return
			}
		}
		done <- nil
	}()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// When a lifecycle command is sent.
	err = Execute(ctx, path, Command{Name: "system_powerdown"})
	// Then asynchronous events do not masquerade as command replies.
	if err != nil {
		t.Fatal(err)
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}
