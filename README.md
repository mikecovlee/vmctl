# vmctl

Configuration-driven Linux and Windows virtual machines on a Linux host,
using Docker, QEMU and KVM. Kylin is an example configuration, not a special
case in the lifecycle engine.

**Development preview.** The JSON contract is `v1alpha1`. See
[compatibility](docs/compatibility.md) for tested guest versions and configurations.

## Requirements

- Linux x86-64 host with `/dev/kvm` and a working Docker daemon.
- Permission to use Docker (equivalent to powerful host access).
- Go 1.25 or newer to build the CLI; `ssh` and `ssh-keygen` for Kickstart.
- Your own OS installation ISO and any required license or driver media.
- Internet access for the initial runtime image build; guest networking is
  isolated by default. Runtime packages are not yet version-locked.

## Quick start

```sh
make build
bin/vmctl image build runtime
bin/vmctl doctor
# Copy an example JSON and set media.iso to your installation media.
bin/vmctl validate my-vm.json
bin/vmctl create my-vm.json
bin/vmctl install my-vm
bin/vmctl console my-vm --show-secret
```

Commands below run from the repository root. `make build` produces `bin/vmctl`;
add that directory to `PATH` to use the shorter `vmctl` commands in the examples.
Build the runtime before `doctor`: its KVM and firmware checks use that image.

For interactive installation, connect a VNC client to the displayed address.
After completing OS installation, shut down the guest, then run:

```sh
bin/vmctl finish-install my-vm
bin/vmctl start my-vm
bin/vmctl status my-vm
bin/vmctl stop my-vm
bin/vmctl delete my-vm --yes
```

`finish-install` records a user declaration, not a successful verification.
For unattended Kylin installation, follow
[examples/kylin-v10](examples/kylin-v10/README.md).
For unattended Ubuntu Server ISO installation and key-only SSH,
follow [examples/ubuntu](examples/ubuntu/README.md).
**Linux memory checkpoints have been verified on Ubuntu 24.04.5, BIOS without
TPM:** checkpoint/restore rolled back a running process's RAM-only value and
counter while retaining its PID and random identity. See the
[validation record](docs/validation.md#linux-memory-checkpoint).
For Windows services and memory-checkpoint experiments, we recommend
[Windows Server 2025](examples/windows-server2025/README.md): BIOS, no TPM,
unattended installation and SSH. For Windows 11 desktop workloads, follow
[examples/windows11](examples/windows11/README.md).
**The Windows 11 UEFI/TPM profile does not support live memory checkpoints.**
It supports pause/resume and stopped snapshots/export/import instead.
The tested Windows 11 build reached the desktop automatically. See the
validation record for installation and SSH results.
The Windows example now prepares offline, key-only OpenSSH by default. Its host
script downloads a pinned MSI, or accepts `OPENSSH_MSI` and `OPENSSH_SHA256` for
offline preparation. The default tested package is an upstream Preview release,
not production-ready. Optional `POST_INSTALL_SCRIPT` and `SOFTWARE_DIR` inputs
provide an offline software-installation hook; see the example for details.

## Design

- Each instance has a private directory, disk, MAC, credentials and QMP socket.
  UEFI NVRAM and TPM state are also per-instance.
- State defaults to `$XDG_STATE_HOME/vmctl` or `~/.local/state/vmctl`.
  Override with `--state-dir PATH` on every command.
- `isolated` uses QEMU user networking with `restrict=on`; IPv6 is disabled.
  Explicit TCP forwards remain reachable. No host service is exposed to the
  guest through a built-in `guestfwd` exception.
- Management listeners default to localhost. Explicit binds allow LAN access.
- `start` is idempotent. `stop` requests ACPI shutdown over QMP and waits up to
  60 seconds. Timeout returns an error and leaves the guest running; only
  `stop --force` cuts power. Containers have no automatic restart policy.
- `verify` performs read-only guest checks for the Kickstart completion marker
  and mounted installation media, then records the verified installation state.
- Structured command output is indented JSON; `--json` selects compact JSON.
  `logs` and `image build` emit runtime text rather than JSON.

See [configuration](docs/configuration.md), [lifecycle](docs/lifecycle.md),
[operations](docs/operations.md), [security](SECURITY.md), and
[contributing](CONTRIBUTING.md). `vmctl version --json` reports build identity
and supported configuration/bundle formats.

## Scope

The [v1.0 release goals](docs/v1.0.md) track single-host lifecycle reliability,
ISO installation, recovery, backup restoration and operational acceptance.
The first release does not provide
clone commands, a web UI, clusters, live migration or a Windows-host backend.
Pause/resume, stopped disk/NVRAM/TPM snapshots and experimental BIOS-only memory
checkpoints are documented in [lifecycle](docs/lifecycle.md).
Stopped instances can also be exported and imported as private bundles with
`export NAME ARCHIVE` and `import ARCHIVE NAME`; see the lifecycle guide for
media relocation, preserved identity and credential handling.
Windows installation and activation remain the user's responsibility.

## Source distribution

`make dist` creates a source archive from an explicit allowlist. It excludes
the legacy Kylin shell implementation, Git history, VM disks, private keys and
installation media. This repository contains only the new CLI and its examples;
legacy files and private development history are not included. No runtime image
is published automatically.

Licensed under Apache-2.0. Runtime packages retain their own licenses; see
[NOTICE](NOTICE). OS media and guest licenses are not included.
