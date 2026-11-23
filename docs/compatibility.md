# Guest compatibility

Requires a Linux x86_64 host, Docker Engine access, and `/dev/kvm`. Windows hosts and ARM hosts are outside this release's scope.

| Example | Installation | Hardware | Verification status |
| --- | --- | --- | --- |
| Linux manual | Interactive ISO | BIOS, VirtIO | Experimental |
| Kylin V10 SP3 | Kickstart | BIOS, VirtIO | Tested: unattended install, disk boot, SSH acceptance, QMP ACPI shutdown |
| Windows manual | Interactive ISO | UEFI, SATA, e1000 | Experimental; no Windows ISO tested |
| Windows 11 manual | Interactive ISO | UEFI, Secure Boot, TPM 2, SATA, e1000 | Experimental; acceptance pending |

Manual installation uses `create`, `install`, then VNC. After installation, `finish-install` records the user's declaration, not a verified operating system. CD-ROM boots first initially, then disk on subsequent boots. `media.extraISOs` supplies optional driver media.

Windows licenses and media are user-supplied. SATA and e1000 avoid initial VirtIO driver requirements. Enabling hardware options does not prove Windows compatibility. Windows automatic installation, Sysprep cloning, and snapshot portability are not included.

Isolated networking restricts outbound traffic while preserving explicit incoming TCP forwards. Egress mode explicitly permits outbound traffic. IPv6 is disabled in this backend. Neither mode creates a shared guest LAN.
