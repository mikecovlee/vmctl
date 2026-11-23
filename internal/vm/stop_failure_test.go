package vm

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"vmctl/internal/config"
	"vmctl/internal/docker"
	"vmctl/internal/state"
)

func TestStopPreservesGuestWhenQMPUnavailable(t *testing.T) {
	testStopPreservesGuest(t, false)
}

func TestStopPreservesGuestWhenShutdownTimesOut(t *testing.T) {
	testStopPreservesGuest(t, true)
}

func testStopPreservesGuest(t *testing.T, acknowledge bool) {
	t.Helper()
	// Given a running managed guest whose monitor is unavailable.
	root := t.TempDir()
	store := state.Store{Root: root}
	if err := store.Save(state.Record{Name: "demo", Phase: state.Verified}); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(root, "demo")
	spec, err := json.Marshal(config.Spec{Name: "demo"})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "spec.json"), spec, 0600); err != nil {
		t.Fatal(err)
	}
	log := filepath.Join(root, "commands")
	binary := filepath.Join(root, "docker")
	script := "#!/bin/sh\nprintf '%s\\n' \"$1\" >> '" + log + "'\ncase \"$1\" in\nps) printf 'container\\n';;\ninspect) printf '%s\\n' '[{\"Id\":\"container\",\"State\":{\"Running\":true},\"Config\":{\"Labels\":{\"io.vmctl.managed\":\"true\",\"io.vmctl.instance\":\"" + dir + "\"}}}]';;\nesac\n"
	if err := os.WriteFile(binary, []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	manager := Manager{Store: store, Docker: docker.Client{Binary: binary}}
	ctx := context.Background()
	if acknowledge {
		listener, err := net.Listen("unix", filepath.Join(dir, "qmp.sock"))
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
			if err := encoder.Encode(json.RawMessage(`{"QMP":{}}`)); err != nil {
				done <- err
				return
			}
			for range 2 {
				var request struct {
					ID string `json:"id"`
				}
				if err := decoder.Decode(&request); err != nil {
					done <- err
					return
				}
				if err := encoder.Encode(struct {
					ID     string   `json:"id"`
					Return struct{} `json:"return"`
				}{ID: request.ID}); err != nil {
					done <- err
					return
				}
			}
			done <- nil
		}()
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Second)
		defer cancel()
		defer func() {
			if err := <-done; err != nil {
				t.Error(err)
			}
		}()
	}
	// When graceful shutdown cannot reach QMP.
	err = manager.Stop(ctx, "demo", false)
	// Then the failure is surfaced without terminating the guest.
	if err == nil {
		t.Fatal("missing monitor accepted")
	}
	if acknowledge && !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline: %v", err)
	}
	commands, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	for _, command := range strings.Fields(string(commands)) {
		if command == "kill" || command == "rm" || command == "stop" {
			t.Fatalf("destructive fallback: %s", commands)
		}
	}
}
