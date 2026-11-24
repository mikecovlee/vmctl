# Ubuntu 24.04 LTS

Ubuntu Server ISO installation to a new empty disk, using the same lifecycle as
the Kylin example: `create`, `install`, `finish-install`, then `start`.
Defaults: 2 vCPUs, 2048 MiB RAM, 20 GiB VirtIO disk, BIOS, no TPM, isolated
networking, SSH on localhost **2260** and VNC on localhost **5960**.

## Prepare and install

Build vmctl and its runtime using the root README and add `bin/` to `PATH`.
Host tools: Bash, `xorriso`, `ssh-keygen`, `sha256sum`, `sed` and `grep`.
Supply an amd64 Ubuntu **24.04 live-server ISO**, with its matching trusted
SHA-256 from [Ubuntu releases](https://releases.ubuntu.com/24.04/).
Desktop and older installation media are not supported by this generator.

Run from the repository root:

```sh
UBUNTU_SHA256='<matching official SHA-256>' \
  bash examples/ubuntu/prepare.sh /path/to/ubuntu-24.04-live-server-amd64.iso
export VMCTL_UBUNTU_STATE="$PWD/examples/ubuntu/generated/state"
vmctl --state-dir "$VMCTL_UBUNTU_STATE" create examples/ubuntu/vm.json
vmctl --state-dir "$VMCTL_UBUNTU_STATE" install ubuntu
```

Preparation generates a private SSH key, embeds `autoinstall.yaml` into a derived
installation ISO and enables unattended boot in its GRUB configuration. `xorriso`
replays the original ISO's boot equipment. Only the public key enters the ISO.
There is one installation ISO; `create` allocates the separate empty system disk.
`install.mode` is `manual` because the media itself drives the unattended installer.

The installer partitions the VM disk, installs the standard Ubuntu Server source
and OpenSSH from the media, adds key-only access for `vmctl`, writes
`/etc/vmctl-install-complete`, and powers off. No guest Internet is required;
installer refresh is disabled and APT permits offline installation. Wait for
`vmctl status ubuntu` (with the same state directory) to report stopped. Then:

```sh
vmctl --state-dir "$VMCTL_UBUNTU_STATE" finish-install ubuntu
vmctl --state-dir "$VMCTL_UBUNTU_STATE" start ubuntu
ssh -p 2260 -i examples/ubuntu/generated/client-key \
  -o IdentitiesOnly=yes -o StrictHostKeyChecking=accept-new \
  -o UserKnownHostsFile=examples/ubuntu/generated/known_hosts \
  vmctl@127.0.0.1 'sudo cloud-init status --wait && test -f /etc/vmctl-install-complete && cat /etc/os-release'
```

Check the installer logs/VNC if it exits before completion. `finish-install`
records a declaration; the SSH marker check supplies the acceptance evidence.
The generic Kickstart `verify` command does not verify this example.

The account has a locked password and passwordless sudo. Protect the private
key and derived media. Preparation refuses an existing `generated/` directory
and retains partial output for diagnosis. That directory also contains instance
state in this walkthrough: do not remove it while retaining the VM. Keep the
ISO available while referenced. Budget for the input ISO, derived ISO and VM disk.
Change ports before creating another instance if these ports are occupied.

## Memory checkpoint

```sh
vmctl --state-dir "$VMCTL_UBUNTU_STATE" checkpoint ubuntu before-change --yes
vmctl --state-dir "$VMCTL_UBUNTU_STATE" restore-checkpoint ubuntu before-change --yes
vmctl --state-dir "$VMCTL_UBUNTU_STATE" stop ubuntu
```

This exact ISO preparation and fresh installation flow was verified on Ubuntu
24.04.5, including SSH and normal shutdown/start. Linux memory rollback was also
verified on that ISO-installed guest: a process retained its PID and random RAM
identity while its value and mutation counter rolled back. Export/import retained
the internal checkpoint and successfully restored it. See the
[validation record](../../docs/validation.md#linux-memory-checkpoint).
Checkpoints remain experimental and require BIOS without TPM; runtime/host
portability and application consistency are not guaranteed. A timeout does not
cancel a QEMU operation. See [lifecycle](../../docs/lifecycle.md).

Generated media, keys and state are ignored and excluded from source archives.
Keep resource-limited guest tests sequential.

Reference: [Canonical autoinstall delivery](https://canonical-subiquity.readthedocs-hosted.com/en/latest/tutorial/providing-autoinstall.html).
