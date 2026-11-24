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
| [Ubuntu 24.04 LTS](../examples/ubuntu/README.md) | Server ISO with embedded autoinstall | BIOS, no TPM, VirtIO; 2 vCPU, 2 GiB, 20 GiB | 24.04.5 ISO installation, SSH, restart, RAM checkpoint rollback and checkpoint retention across export/import verified |
| Kylin V10 SP3 | Kickstart | BIOS, VirtIO | Tested: unattended install, disk boot, SSH acceptance, QMP ACPI shutdown |
| [Windows Server 2025](../examples/windows-server2025/README.md) (recommended for services/checkpoints) | BIOS answer media and offline SSH | BIOS, no TPM, SATA, e1000 | Standard Evaluation 26100: zero-input installation, SSH and RAM-only rollback tested; subsequent CLI save/restore passed with default 5m wait; first stop after restore timed out, later normal stop passed |
| Windows manual | Interactive ISO | UEFI, SATA, e1000 | Experimental; general Windows versions not tested |
| [Windows 11](../examples/windows11/README.md) | Derived no-prompt UEFI ISO with answer file | 2 vCPU, 4096 MiB, 64 GiB; UEFI, Secure Boot, TPM 2, SATA, e1000 | Pro 26300.9457: zero-input desktop installation; separate prototype verified SSH, restart and guest Secure Boot/TPM readiness; final repository scripts not fully redeployed; activation and RDP unverified |

Manual installation uses `create`, `install`, then VNC. After installation, `finish-install` records the user's declaration, not a verified operating system. CD-ROM boots first initially, then disk on subsequent boots. `media.extraISOs` supplies optional driver media.

**Windows 11: live memory checkpoints are not supported by vmctl's current
UEFI/TPM profile.** `checkpoint` and `restore-checkpoint` reject this configuration.
Pause/resume and stopped snapshots/export/import remain available. Use the BIOS,
no-TPM Server 2025 example for experimental memory checkpoints.

Linux memory checkpoints have also been verified on Ubuntu 24.04.5 using BIOS
without TPM: process identity and PID were preserved while RAM-only values
rolled back. This does not imply every Linux distribution or application has
been tested. See the [checkpoint evidence](validation.md#linux-memory-checkpoint).

Windows licenses and media are user-supplied. SATA and e1000 avoid initial VirtIO driver requirements. Enabling hardware options does not prove guest-side Secure Boot or TPM validation. The unattended example is tested only on the build listed above; automatic Windows verification, Sysprep cloning, and snapshot portability are not included.

Isolated networking restricts outbound traffic while preserving explicit incoming TCP forwards. NAT mode (`nat`, legacy alias `egress`) explicitly permits outbound traffic. Incoming forwards support TCP only, not UDP. IPv6 is disabled in this backend. Neither mode creates a shared guest LAN. Modes and forwards can be changed with `vmctl network` while the instance is stopped.
