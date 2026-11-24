# Guest compatibility

The Windows 11 example includes offline OpenSSH bootstrap. A separate zero-input
trial on Pro 26300.9457 verified key-only SSH, automatic sshd startup and reconnect
after normal shutdown/start with the same host key. The tested MSI is
10.4.0.0p1-Preview (upstream labels it non-production-ready). Preparation supports
a pinned host download or a supplied local MSI; neither requires guest Internet
access. See [validation](validation.md) for the remaining test boundaries.

Requires a Linux x86_64 host, Docker Engine access, and `/dev/kvm`. Windows hosts and ARM hosts are outside this release's scope.

| Example | Installation | Hardware | Verification status |
| --- | --- | --- | --- |
| Linux manual | Interactive ISO | BIOS, VirtIO | Experimental |
| Kylin V10 SP3 | Kickstart | BIOS, VirtIO | Tested: unattended install, disk boot, SSH acceptance, QMP ACPI shutdown |
| Windows manual | Interactive ISO | UEFI, SATA, e1000 | Experimental; general Windows versions not tested |
| [Windows 11](../examples/windows11/README.md) | Derived no-prompt UEFI ISO with answer file | 2 vCPU, 4096 MiB, 64 GiB; UEFI, Secure Boot, TPM 2, SATA, e1000 | Pro 26300.9457: one fresh install reached desktop with zero keyboard/mouse input; Windows acceptance pending |

Manual installation uses `create`, `install`, then VNC. After installation, `finish-install` records the user's declaration, not a verified operating system. CD-ROM boots first initially, then disk on subsequent boots. `media.extraISOs` supplies optional driver media.

Windows licenses and media are user-supplied. SATA and e1000 avoid initial VirtIO driver requirements. Enabling hardware options does not prove guest-side Secure Boot or TPM validation. The unattended example is tested only on the build listed above; automatic Windows verification, Sysprep cloning, and snapshot portability are not included.

Isolated networking restricts outbound traffic while preserving explicit incoming TCP forwards. Egress mode explicitly permits outbound traffic. IPv6 is disabled in this backend. Neither mode creates a shared guest LAN.
