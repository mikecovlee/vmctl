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

func TestLongCommandWaitsBeyondOrdinaryDeadline(t *testing.T) {
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
		encoder, decoder := json.NewEncoder(conn), json.NewDecoder(conn)
		if err := encoder.Encode(json.RawMessage(`{"QMP":{}}`)); err != nil {
			done <- err
			return
		}
		for n := range 2 {
			var request struct{ ID string }
			if err := decoder.Decode(&request); err != nil {
				done <- err
				return
			}
			if n == 1 {
				timer := time.NewTimer(11 * time.Second)
				defer timer.Stop()
				<-timer.C
			}
			if err := encoder.Encode(struct {
				ID     string `json:"id"`
				Return string `json:"return"`
			}{request.ID, "saved"}); err != nil {
				done <- err
				return
			}
		}
		done <- nil
	}()
	result, err := Query(context.Background(), path, Command{Name: "human-monitor-command", Timeout: 20 * time.Second})
	if err != nil || string(result) != `"saved"` {
		t.Fatalf("long operation: result=%s err=%v", result, err)
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

func TestDisconnectedCommandHasUnknownOutcome(t *testing.T) {
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
		encoder, decoder := json.NewEncoder(conn), json.NewDecoder(conn)
		if err := encoder.Encode(json.RawMessage(`{"QMP":{}}`)); err != nil {
			done <- err
			return
		}
		var req struct{ ID string }
		if err := decoder.Decode(&req); err != nil {
			done <- err
			return
		}
		if err := encoder.Encode(json.RawMessage(`{"id":"vmctl-0","return":{}}`)); err != nil {
			done <- err
			return
		}
		if err := decoder.Decode(&req); err != nil {
			done <- err
			return
		}
		done <- nil
	}()
	_, err = Query(context.Background(), path, Command{Name: "human-monitor-command"})
	if !errors.Is(err, ErrOutcomeUnknown) {
		t.Fatalf("expected unknown outcome after dispatch, got %v", err)
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}
