package docker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

var ErrCommand = errors.New("docker command failed")
var ErrOwnership = errors.New("container is not owned by vmctl")

type Client struct {
	Binary  string
	Timeout time.Duration
}

func (c Client) Exists(ctx context.Context, instance string) (bool, error) {
	out, err := c.Run(ctx, "ps", "--all", "--filter", "name=^/"+ContainerName(instance)+"$", "--format", "{{.ID}}")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(out)) != "", nil
}

func (c Client) Run(ctx context.Context, args ...string) ([]byte, error) {
	binary := c.Binary
	if binary == "" {
		binary = "docker"
	}
	limit := c.Timeout
	if limit == 0 {
		limit = 30 * time.Second
	}
	bounded, cancel := context.WithTimeout(ctx, limit)
	defer cancel()
	cmd := exec.CommandContext(bounded, binary, args...)
	cmd.WaitDelay = time.Second
	out, err := cmd.CombinedOutput()
	if bounded.Err() != nil {
		return nil, bounded.Err()
	}
	if err != nil {
		return nil, fmt.Errorf("%w: %w: %s", ErrCommand, err, out)
	}
	return out, nil
}

type Container struct {
	ID    string `json:"Id"`
	State struct {
		Running  bool `json:"Running"`
		ExitCode int  `json:"ExitCode"`
	} `json:"State"`
	Config struct {
		Labels map[string]string `json:"Labels"`
	} `json:"Config"`
}

func (c Client) Inspect(ctx context.Context, name string) (Container, error) {
	out, err := c.Run(ctx, "inspect", name)
	if err != nil {
		return Container{}, err
	}
	var containers []Container
	if err := json.Unmarshal(out, &containers); err != nil {
		return Container{}, fmt.Errorf("inspect: %w", err)
	}
	if len(containers) != 1 {
		return Container{}, fmt.Errorf("%w: expected one container", ErrCommand)
	}
	if containers[0].Config.Labels["io.vmctl.managed"] != "true" {
		return Container{}, ErrOwnership
	}
	return containers[0], nil
}

func (c Client) InspectInstance(ctx context.Context, instance string) (Container, error) {
	container, err := c.Inspect(ctx, ContainerName(instance))
	if err != nil {
		return Container{}, err
	}
	if container.Config.Labels["io.vmctl.instance"] != instance {
		return Container{}, ErrOwnership
	}
	return container, nil
}
