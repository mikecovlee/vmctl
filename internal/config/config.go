package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/netip"
	"regexp"
)

var ErrInvalid = errors.New("invalid VM configuration")

type Spec struct {
	APIVersion string    `json:"apiVersion"`
	Name       string    `json:"name"`
	Resources  Resources `json:"resources"`
	Media      Media     `json:"media"`
	Machine    Machine   `json:"machine"`
	Install    Install   `json:"install"`
	Network    Network   `json:"network"`
	Console    Console   `json:"console"`
}
type Resources struct {
	CPU       int `json:"cpu"`
	MemoryMiB int `json:"memoryMiB"`
	DiskGiB   int `json:"diskGiB"`
}
type Media struct {
	ISO       string   `json:"iso"`
	ExtraISOs []string `json:"extraISOs"`
}
type Machine struct {
	Firmware   string `json:"firmware"`
	SecureBoot bool   `json:"secureBoot"`
	TPM        bool   `json:"tpm"`
	DiskBus    string `json:"diskBus"`
	NIC        string `json:"nic"`
}
type Install struct {
	Mode       string `json:"mode"`
	Template   string `json:"template"`
	KernelPath string `json:"kernelPath"`
	InitrdPath string `json:"initrdPath"`
}
type Network struct {
	Mode     string    `json:"mode"`
	Forwards []Forward `json:"forwards"`
}
type Forward struct {
	Bind  string `json:"bind"`
	Host  int    `json:"host"`
	Guest int    `json:"guest"`
}
type Console struct {
	VNCBind string `json:"vncBind"`
	VNCPort int    `json:"vncPort"`
}

var namePattern = regexp.MustCompile(`^[a-z][a-z0-9-]{0,47}$`)

func Parse(r io.Reader) (Spec, error) {
	s := Spec{Machine: Machine{Firmware: "bios", DiskBus: "virtio", NIC: "virtio-net-pci"}, Install: Install{Mode: "manual"}, Network: Network{Mode: "isolated"}, Console: Console{VNCBind: "127.0.0.1", VNCPort: 5901}}
	data, err := io.ReadAll(io.LimitReader(r, (1<<20)+1))
	if err != nil {
		return Spec{}, fmt.Errorf("%w: %v", ErrInvalid, err)
	}
	if len(data) > 1<<20 {
		return Spec{}, fmt.Errorf("%w: configuration exceeds 1 MiB", ErrInvalid)
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if err := d.Decode(&s); err != nil {
		return Spec{}, fmt.Errorf("%w: %v", ErrInvalid, err)
	}
	var trailing json.RawMessage
	if err := d.Decode(&trailing); err != io.EOF {
		return Spec{}, fmt.Errorf("%w: trailing input", ErrInvalid)
	}
	for i := range s.Network.Forwards {
		if s.Network.Forwards[i].Bind == "" {
			s.Network.Forwards[i].Bind = "127.0.0.1"
		}
	}
	if s.Network.Mode == "egress" {
		s.Network.Mode = "nat"
	}
	if err := s.validate(); err != nil {
		return Spec{}, err
	}
	return s, nil
}

func (s Spec) validate() error {
	bad := func(field string) error { return fmt.Errorf("%w: %s", ErrInvalid, field) }
	if s.APIVersion != "v1alpha1" || !namePattern.MatchString(s.Name) {
		return bad("apiVersion/name")
	}
	if s.Resources.CPU < 1 || s.Resources.CPU > 256 || s.Resources.MemoryMiB < 256 || s.Resources.DiskGiB < 1 {
		return bad("resources")
	}
	if s.Media.ISO == "" {
		return bad("media.iso")
	}
	switch s.Network.Mode {
	case "isolated", "nat":
	default:
		return bad("network.mode")
	}
	switch s.Machine.Firmware {
	case "bios", "uefi":
	default:
		return bad("machine.firmware")
	}
	if (s.Machine.SecureBoot || s.Machine.TPM) && s.Machine.Firmware != "uefi" {
		return bad("UEFI required for Secure Boot/TPM")
	}
	switch s.Machine.DiskBus {
	case "virtio", "sata":
	default:
		return bad("machine.diskBus")
	}
	switch s.Machine.NIC {
	case "virtio-net-pci", "e1000":
	default:
		return bad("machine.nic")
	}
	switch s.Install.Mode {
	case "manual":
	case "kickstart":
		if s.Install.Template == "" || s.Install.KernelPath == "" || s.Install.InitrdPath == "" {
			return bad("kickstart template/kernelPath/initrdPath")
		}
	default:
		return bad("install.mode")
	}
	vncAddress, err := netip.ParseAddr(s.Console.VNCBind)
	if err != nil || !vncAddress.Is4() || s.Console.VNCPort < 5900 || s.Console.VNCPort > 65535 {
		return bad("console")
	}
	ports := map[int]bool{s.Console.VNCPort: true}
	for _, f := range s.Network.Forwards {
		address, err := netip.ParseAddr(f.Bind)
		if err != nil || !address.Is4() || f.Host < 1 || f.Host > 65535 || f.Guest < 1 || f.Guest > 65535 || ports[f.Host] {
			return bad("network.forwards")
		}
		ports[f.Host] = true
	}
	return nil
}
