# Validation record

## Linux memory checkpoint

### Repository ISO installation and RAM rollback

Linux memory checkpoint/restore was verified using the repository's Ubuntu
**24.04.5 LTS** single-ISO autoinstall example: BIOS, no TPM, 2 vCPUs, 2048 MiB
RAM, a 20 GiB VirtIO disk and isolated networking. The official live-server ISO
matched the publisher's SHA-256:
`97f3d7ffb032c3eb3b23d2c8be9cc76e60c2c1f2c0146ba5ba9fe01cafae0fd8`.

The repository preparation script embedded autoinstall configuration into a
derived ISO. `create` allocated a new empty qcow2 disk; `install` booted the ISO
and the Ubuntu installer partitioned and installed that disk. No preinstalled
disk was imported or substituted. Installation ran offline, powered off
automatically, and booted from disk with key-only SSH, the installation marker
and a 20 GiB root partition.
Normal shutdown/start preserved SSH host identity. A temporary Bash process then
retained PID 1583 and random identity `e42dcb8d-e293-4b20-8206-28bd33325b21`
while its RAM-only value/counter changed `before/0` → `after/1` → `before/0`.
Both checkpoint CLI operations succeeded with the new five-minute default.

A stopped snapshot and export/import preserved the internal `iso-ram` checkpoint
(953 MiB of saved VM state). Actual restore in the imported instance recovered
the same process identity and `before/0` values. Test scripts/services were removed
and guests stopped afterward; retained checkpoints can restore those artifacts.

Windows Server 2025 BIOS/no-TPM checkpoint save and restore also returned success
with the new default. The first normal stop immediately after restore timed out;
a subsequent normal stop succeeded. No forced power-off was used.

On the ISO-installed Ubuntu guest, controlled host TCP and UDP endpoints were
first verified reachable from the host, then blocked in isolated mode and reachable
in egress mode. Both guest probes confirmed no IPv6 default route. The disposable
instance configuration was returned to isolated mode and the instance stopped.

### Historical cloud-image trial

An earlier, separate Ubuntu 24.04.5 trial used an official Noble cloud image
and NoCloud with the same BIOS/no-TPM resource profile. Cloud-init, key-only SSH
and normal shutdown/start passed. A temporary Python service retained its PID
1271 and random identity as its RAM-only value/counter changed
`before/0` → `after/1` → `before/0`. Both checkpoint CLI commands succeeded.
The service and script were removed afterward; its retained checkpoint can
restore them. The repository ISO installation results are recorded above.

## Command verification coverage

The table distinguishes real guest acceptance from automated test coverage.

| Surface | Real guest evidence | Notes |
| --- | --- | --- |
| `create`, `install`, `finish-install`, `start` | Fresh Ubuntu repository ISO installation; earlier Kylin and Windows trials | Windows final-script boundaries are recorded below |
| `list`, `status`, SSH and normal `stop`/`start` | Used during guest acceptance | Server's first stop after checkpoint restore timed out; subsequent normal stop passed |
| `checkpoint`, `restore-checkpoint` | Ubuntu and Server RAM rollback; latest CLI save/restore passed with default 5m wait | BIOS without TPM; custom timeout parsing covered by automated tests |
| `snapshot`, `snapshots`, `restore` | Ubuntu stopped snapshot created and its internal checkpoint inspected | Automated snapshot/restore coverage; stopped `restore` was not rerun end to end in the latest trial |
| `export`, `import` | Ubuntu exported, imported, booted and restored an internal RAM checkpoint | Same host/runtime, not cross-host compatibility evidence |
| Network isolation and NAT | Controlled TCP/UDP blocked in isolated mode and reachable in egress mode | Latest trial changed disposable configuration directly; network mutation CLI commands covered by automated tests |
| `pause`, `resume` | Not rerun in latest guest acceptance | Covered by automated tests |
| `console`, `logs` | Used during installation trials | Not every listener binding or logging condition was tested |
| `stop --force`, `delete` | Latest trial used no forced power-off; earlier Kylin instances were deleted via CLI | Latest instances retained stopped |
| `verify` | Kylin read-only SSH acceptance passed | Kickstart-specific; not an Ubuntu or Windows verifier |

Tests ran on a Linux Docker/KVM host. Existing legacy VM storage was not modified.

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

The instance was marked `installed-unverified` using `finish-install`.
SSH, Secure Boot/TPM and shutdown/start results from the subsequent trial are
recorded above. Activation and RDP were not tested.
The generator in `examples/windows11` packages this tested
procedure; generated credentials and ISO images are excluded from source archives.
