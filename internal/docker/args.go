package docker

import (
	"fmt"
	"strconv"
	"vmctl/internal/config"
)

func QEMUArgs(s config.Spec, mac string, installing bool) []string {
	bus := "virtio"
	if s.Machine.DiskBus == "sata" {
		bus = "ide"
	}
	network := "user,id=net0,ipv6=off"
	if s.Network.Mode == "isolated" {
		network += ",restrict=on"
	}
	for _, f := range s.Network.Forwards {
		network += fmt.Sprintf(",hostfwd=tcp:%s:%d-:%d", f.Bind, f.Host, f.Guest)
	}
	args := []string{"-name", s.Name, "-machine", "q35", "-accel", "kvm", "-cpu", "host",
		"-m", strconv.Itoa(s.Resources.MemoryMiB), "-smp", strconv.Itoa(s.Resources.CPU),
		"-drive", "file=/instance/disk.qcow2,format=qcow2,if=" + bus,
		"-cdrom", "/media/install.iso", "-netdev", network,
		"-device", s.Machine.NIC + ",netdev=net0,mac=" + mac,
		"-display", "none", "-vnc", fmt.Sprintf("%s:%d,password=on", s.Console.VNCBind, s.Console.VNCPort-5900),
		"-serial", "stdio", "-qmp", "unix:/instance/qmp.sock,server=on,wait=off"}
	for i := range s.Media.ExtraISOs {
		args = append(args, "-drive", fmt.Sprintf("file=/media/extra-%d.iso,media=cdrom,readonly=on", i))
	}
	if s.Machine.Firmware == "uefi" {
		code := "/usr/share/OVMF/OVMF_CODE_4M.fd"
		if s.Machine.SecureBoot {
			code = "/usr/share/OVMF/OVMF_CODE_4M.secboot.fd"
			args = append(args, "-global", "driver=cfi.pflash01,property=secure,value=on", "-global", "ICH9-LPC.disable_s3=1", "-machine", "q35,smm=on")
		}
		args = append(args, "-drive", "if=pflash,format=raw,readonly=on,file="+code, "-drive", "if=pflash,format=raw,file=/instance/nvram.fd")
	}
	if s.Machine.TPM {
		args = append(args, "-chardev", "socket,id=chrtpm,path=/instance/tpm/socket", "-tpmdev", "emulator,id=tpm0,chardev=chrtpm", "-device", "tpm-tis,tpmdev=tpm0")
	}
	switch {
	case installing && s.Install.Mode == "kickstart":
		args = append(args, "-kernel", "/instance/boot/vmlinuz", "-initrd", "/instance/boot/initrd-ks.img", "-append", "inst.ks=file:/ks.cfg inst.repo=cdrom console=ttyS0,115200n8")
	case installing:
		args = append(args, "-boot", "order=c,once=d")
	default:
		args = append(args, "-boot", "order=c")
	}
	return args
}
