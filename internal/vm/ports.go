package vm

import (
	"fmt"
	"net"
	"strconv"

	"vmctl/internal/config"
)

func checkPorts(spec config.Spec) error {
	addresses := []string{net.JoinHostPort(spec.Console.VNCBind, strconv.Itoa(spec.Console.VNCPort))}
	for _, forward := range spec.Network.Forwards {
		addresses = append(addresses, net.JoinHostPort(forward.Bind, strconv.Itoa(forward.Host)))
	}
	for _, address := range addresses {
		listener, err := net.Listen("tcp", address)
		if err != nil {
			return fmt.Errorf("port %s unavailable: %w", address, err)
		}
		if err := listener.Close(); err != nil {
			return err
		}
	}
	return nil
}
