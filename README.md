# vmctl

Configuration-driven Linux and Windows virtual machines on a Linux host,
using Docker, QEMU and KVM. Kylin is an example configuration, not a special
case in the lifecycle engine.

**Development preview.** The JSON contract is `v1alpha1`. Windows examples
have build-specific evidence, not blanket compatibility guarantees. See
[compatibility](docs/compatibility.md) for evidence and limitations.

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
For the tested zero-input Windows 11 Pro installation, follow
[examples/windows11](examples/windows11/README.md).
Its single tested build reached the desktop automatically; Windows activation
and comprehensive guest verification are separate and remain untested.

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
  It is not a comprehensive OS or network security audit.
- Structured command output is indented JSON; `--json` selects compact JSON.
  `logs` and `image build` emit runtime text rather than JSON.

See [configuration](docs/configuration.md), [lifecycle](docs/lifecycle.md),
[security](SECURITY.md), and [contributing](CONTRIBUTING.md).

## Scope

The first release focuses on single-host lifecycle management, manual ISO
installation and a generic Kickstart template renderer. It does not provide
clone commands, a web UI, clusters, live migration or a Windows-host backend.
Pause/resume, stopped disk/NVRAM/TPM snapshots and experimental BIOS-only memory
checkpoints are documented in [lifecycle](docs/lifecycle.md).
Windows installation and activation remain the user's responsibility.

## Source distribution

`make dist` creates a source archive from an explicit allowlist. It excludes
the legacy Kylin shell implementation, Git history, VM disks, private keys and
installation media. This repository contains only the new CLI and its examples;
legacy files and private development history are not included. No runtime image
is published automatically.

Licensed under Apache-2.0. Runtime packages retain their own licenses; see
[NOTICE](NOTICE). OS media and guest licenses are not included.
