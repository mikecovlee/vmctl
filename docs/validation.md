# Validation record

## Linux memory checkpoint

Linux memory checkpoint/restore was verified on Ubuntu **24.04.5 LTS**, initialized from the official Noble cloud image using
NoCloud, with BIOS, no TPM, 2 vCPUs, 2048 MiB RAM, a 20 GiB VirtIO
disk and isolated networking. Cloud-init completed, key-only SSH worked, and a
normal ACPI shutdown/start preserved the guest and SSH host identity.

A temporary systemd service ran a Python process whose random identity, value
and mutation counter existed only in RAM. Each observation used a fresh SSH
connection. The results were:

| Observation | PID | RAM value | Mutation counter |
| --- | --- | --- | --- |
| Before `checkpoint` | 1271 | `before` | 0 |
| After mutation | 1271 | `after` | 1 |
| After `restore-checkpoint` | 1271 | `before` | 0 |

The random process identity remained identical across all three observations.
Both vmctl checkpoint commands returned successfully without a timeout. This
demonstrates Linux process-memory rollback, not merely restoration of a file.
The test service and script were removed afterward; restoring the retained
checkpoint would bring those artifacts back.

The repository's single-ISO autoinstall example was subsequently executed against
Ubuntu 24.04.5 live-server media (SHA-256
`97f3d7ffb032c3eb3b23d2c8be9cc76e60c2c1f2c0146ba5ba9fe01cafae0fd8`).
It installed to a fresh empty disk offline, powered off automatically, and booted
from disk with key-only SSH, the installation marker and a 20 GiB root partition.
Normal shutdown/start preserved SSH host identity. A temporary Bash process then
retained PID 1583 and random identity `e42dcb8d-e293-4b20-8206-28bd33325b21`
while its RAM-only value/counter changed `before/0` → `after/1` → `before/0`.
Both checkpoint CLI operations succeeded with the new five-minute default.

A stopped snapshot and export/import preserved the internal `iso-ram` checkpoint
(953 MiB of saved VM state). Actual restore in the imported instance recovered
the same process identity and `before/0` values. Test scripts/services were removed
and guests stopped afterward; retained checkpoints can restore those artifacts.
Other Linux versions, database consistency and cross-host/runtime portability
remain unverified.

Windows Server 2025 BIOS/no-TPM checkpoint save and restore also returned success
with the new default. The first normal stop immediately after restore timed out;
a subsequent normal stop succeeded. No forced power-off was used.

On the ISO-installed Ubuntu guest, controlled host TCP and UDP endpoints were
first verified reachable from the host, then blocked in isolated mode and reachable
in egress mode. Both guest probes confirmed no IPv6 default route. The disposable
instance configuration was returned to isolated mode and the instance stopped.

This is a foundation release candidate, not a claim that every guest OS is
certified. Tests ran on a Linux Docker/KVM host using the supplied Kylin V10
x86-64 installation media. Existing legacy VM storage was not modified.

## Executed

### Windows offline SSH bootstrap follow-up

A separate fresh Pro 26300.9457 instance installed with zero keyboard/mouse input
and reached key-only SSH using a bundled Microsoft-signed OpenSSH
10.4.0.0p1-Preview x64 MSI. The trial verified its SHA-256 and Authenticode
signature, automatic/running sshd, disabled password and keyboard-interactive
authentication, administrator authorized-key ACLs, and SSH reconnect after normal
shutdown/start with strict host-key checking. Guest checks also reported Secure
Boot enabled and TPM ready. This does not prove activation or remote attestation.

That trial used a private prototype. The subsequent repository bootstrap and
software-hook example passed Windows PowerShell parser checks; preparation passed
syntax, checksum-rejection, offline/no-download and download-failure cleanup tests.
The new automatic download success path has not been re-tested end to end, and
the final repository scripts have not undergone another full fresh installation.
No third-party software package has been installed through the optional hook in
an end-to-end test. These are remaining validation gaps, not passing results.

### Core lifecycle

- `go test -race -shuffle=on ./...` and `go vet ./...` passed.
- A dedicated Kylin instance completed unattended installation, disk boot and
  SSH acceptance. Guest identity was Kylin Linux Advanced Server V10 (Halberd),
  memory total was 3932 MiB, the only enabled repository was `vmctl-media`, and
  `/dev/sr0` was mounted read-only at `/mnt/iso`.
- QMP ACPI shutdown completed with guest unmount and power-off messages.
  The opt-in `TestKVMGracefulLifecycle` passed. Repeated start was idempotent.
- Two distinct instances ran concurrently. Test instances were subsequently
  stopped and deleted through the CLI.
- Controlled host TCP, UDP and DNS endpoints were reachable in egress mode
  and blocked in isolated mode. Host-side control probes succeeded first.
  The guest had no IPv6 default route; QEMU disables guest IPv6 networking.
- A UEFI/Secure Boot/TPM configuration launched successfully with SATA/e1000.
  This is a firmware startup smoke test, not a Windows installation test.
- Protocol error, QMP timeout, occupied port, state corruption, cancellation,
  disk creation rollback and stop-without-forced-kill tests passed.
- The source archive was inspected: legacy scripts, Git history, credentials,
  ISO images and VM disks were absent.

## Limits of this evidence

- Windows installation evidence is scoped to the build and procedure recorded
  below; other Windows builds remain experimental.
- The DNS test exercises a controlled UDP DNS endpoint, not every behavior of
  SLIRP's built-in DNS resolver. IPv6 has a route/configuration check rather
  than an enabled-IPv6 egress comparison.
- Stop timeout preservation is tested with a protocol peer and runtime fake;
  a real unresponsive operating system was not deliberately induced.
- `verify` checks the installation marker and media mount using read-only SSH.
  OS, memory and repository checks above were separate acceptance commands.
- Independent review was not run: execution was performed directly, as
  requested. Source was published to https://github.com/mikecovlee/vmctl
  after user authorization. No release artifact or version tag was published.

Run ordinary tests with `make test`. KVM integration tests require explicit
`VMCTL_KVM_TEST=1`, `VMCTL_TEST_STATE` and `VMCTL_TEST_INSTANCE` settings;
network tests additionally require `VMCTL_TEST_SSH_PORT`. Use disposable,
dedicated instances, never production data.

## Windows 11 zero-input installation

Windows 11 Pro **26300.9457**, image index **4**, completed one fresh
installation to the desktop with no VNC keyboard/mouse or QMP key injection.
The instance used 2 vCPUs, 4096 MiB RAM, a new 64 GiB disk, UEFI/Secure Boot/TPM
options, SATA/e1000 and isolated networking. Existing instances were untouched.

The derived ISO selected the original media's `efisys_noprompt.bin` as its UEFI
boot image and included an answer template with a matching
Pro default setup key and `WillShowUI=Never`. Passive QMP screenshots showed
firmware loading, Setup at 22%, the post-reboot installation phase, and finally
the desktop. Neither the DVD prompt nor product-key page required input.
Offline OOBE and automatic local-account login completed.

This was one run, not a repeated-install reliability test. The instance was
marked `installed-unverified` using `finish-install`. Windows activation,
guest-side Secure Boot/TPM acceptance, RDP, and Windows ACPI shutdown were not
verified. The generator in `examples/windows11` packages this tested
procedure; generated credentials and ISO images are excluded from source archives.
