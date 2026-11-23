package qmp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"time"
)

var ErrProtocol = errors.New("QMP protocol error")

type Command struct {
	Name      string
	Arguments json.RawMessage
}

type response struct {
	Greeting json.RawMessage `json:"QMP"`
	ID       string          `json:"id"`
	Event    string          `json:"event"`
	Return   json.RawMessage `json:"return"`
	Error    *struct {
		Class       string `json:"class"`
		Description string `json:"desc"`
	} `json:"error"`
}

func Execute(ctx context.Context, path string, command Command) error {
	bounded, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	conn, err := (&net.Dialer{}).DialContext(bounded, "unix", path)
	if err != nil {
		return fmt.Errorf("connect QMP: %w", err)
	}
	defer conn.Close()
	deadline, ok := bounded.Deadline()
	if !ok {
		return fmt.Errorf("%w: missing deadline", ErrProtocol)
	}
	if err := conn.SetDeadline(deadline); err != nil {
		return err
	}
	stop := context.AfterFunc(bounded, func() { conn.Close() })
	defer stop()
	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)
	var greeting response
	if err := decoder.Decode(&greeting); err != nil {
		return fmt.Errorf("QMP greeting: %w", err)
	}
	if len(greeting.Greeting) == 0 {
		return fmt.Errorf("%w: missing greeting", ErrProtocol)
	}
	for index, cmd := range []Command{{Name: "qmp_capabilities"}, command} {
		id := fmt.Sprintf("vmctl-%d", index)
		request := struct {
			Execute   string          `json:"execute"`
			Arguments json.RawMessage `json:"arguments,omitempty"`
			ID        string          `json:"id"`
		}{cmd.Name, cmd.Arguments, id}
		if err := encoder.Encode(request); err != nil {
			return err
		}
		for {
			var reply response
			if err := decoder.Decode(&reply); err != nil {
				return fmt.Errorf("QMP %s: %w", cmd.Name, err)
			}
			if reply.Event != "" {
				continue
			}
			if reply.ID != id {
				return fmt.Errorf("%w: unexpected response id", ErrProtocol)
			}
			if reply.Error != nil {
				return fmt.Errorf("%w: %s: %s", ErrProtocol, reply.Error.Class, reply.Error.Description)
			}
			if len(reply.Return) == 0 {
				return fmt.Errorf("%w: missing return", ErrProtocol)
			}
			break
		}
	}
	return nil
}
