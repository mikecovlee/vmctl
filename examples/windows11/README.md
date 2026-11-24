# Windows 11

**Live memory checkpoints are not supported for this Windows 11 profile.**
Its UEFI/TPM configuration is rejected by `checkpoint` and `restore-checkpoint`.
Pause/resume and stopped snapshots/export/import remain supported; pausing alone
does not save RAM to disk. For memory-checkpoint experiments, use the
[Windows Server 2025 BIOS profile](../windows-server2025/README.md).

One fresh installation of **Windows 11 Pro 26300.9457, WIM index 4**
reached the desktop with **no keyboard or mouse input**, including across Setup
reboots and offline OOBE. Resources: 2 vCPUs, 4096 MiB RAM, 64 GiB disk;
UEFI/Secure Boot/TPM 2 options, SATA, e1000 and isolated networking.

This example includes its own answer template. The generator adds a Pro
default installation key and builds a derived ISO whose UEFI El Torito entry
uses the supplied media's `efi/microsoft/boot/efisys_noprompt.bin`.
No firmware signature patches or hardware-check bypasses are applied.
The default setup key **does not activate Windows**.

## Prepare and run

First build vmctl and its runtime as described in the root README, and put
`bin/` on `PATH`. Run the following commands from the repository root.
On a Linux host, install `7z`, `openssl`, `genisoimage`, `ssh-keygen`, `sha256sum`,
and `wiminfo` (wimlib). Automatic OpenSSH download additionally requires `curl`
and host access to GitHub; Windows itself does not need Internet access.
Supply your own
appropriately licensed ISO. Confirm with `wiminfo` that image index 4 is Pro
and that its language is zh-CN. Other editions, channels, indexes or builds
require adapting and testing the template first.

```sh
bash examples/windows11/prepare.sh /path/to/windows11.iso
vmctl create examples/windows11/vm.json
vmctl install windows11
vmctl console windows11 --show-secret
```

Observe VNC at `127.0.0.1:5904` without sending keys. Setup partitions guest
disk 0 (wiping it), installs Pro, creates local administrator `vmctl`, completes
offline OOBE and requests one automatic login. A single elevated bootstrap then
installs OpenSSH offline, configures key-only authentication and automatic sshd
startup, and allows inbound guest TCP 22. The host exposes it only on localhost:2225.
Without `OPENSSH_MSI`, the host downloads the pinned Microsoft x64 MSI from GitHub
(with a fixed release-asset API fallback) and checks its pinned SHA-256 before
extracting the Windows ISO. Temporary downloads are removed on exit. Both the
host-side hash and guest-side Authenticode signature are checked; no guest download
is needed. The default tested MSI is **10.4.0.0p1-Preview**, explicitly non-production-ready
upstream, SHA-256 `788d8643464a01eb63640a2fd746f9ed23d106b5a901a35f88dc8cd40e446ac9`.
The script deliberately does not download an unpinned latest release.

For an offline host or a different Microsoft-signed package, supply a local MSI
and its trusted publisher checksum. This mode never downloads OpenSSH:

```sh
OPENSSH_MSI=/path/to/OpenSSH-Win64.msi \
OPENSSH_SHA256='replace-with-trusted-sha256' \
bash examples/windows11/prepare.sh /path/to/windows11.iso
```

Both local-package variables are required together. A failed download or checksum
mismatch stops preparation; it never falls back to an unverified package.

After SSH is ready, verify the completion marker before declaring installation:

```sh
ssh -p 2225 -i examples/windows11/generated/client-key \
  -o IdentitiesOnly=yes -o StrictHostKeyChecking=accept-new \
  -o UserKnownHostsFile=examples/windows11/generated/known_hosts \
  vmctl@127.0.0.1 'Get-Service sshd; Get-Content C:\vmctl-install-complete.txt'
```

Startup can take several minutes. Use bounded retries; a listening forwarded port
alone does not mean sshd is ready. `C:\vmctl-ssh-ready.txt` signals SSH setup only;
`C:\vmctl-install-complete.txt` additionally requires the optional software hook
to succeed. Logs are `C:\vmctl-bootstrap.log` and `C:\vmctl-openssh-msi.log`;
failures are recorded in `C:\vmctl-bootstrap-error.txt`.

After successful checks:

```sh
vmctl finish-install windows11
vmctl status windows11
```

`install.mode` remains `manual`: vmctl boots the media; the answer file drives
Windows Setup. `finish-install` is still a user declaration, not automatic
Windows verification. Initial boot is DVD-first, subsequent boots disk-first;
do not repeatedly force DVD boot on an installed disk.

## Optional offline software

Set `POST_INSTALL_SCRIPT=/path/to/post-install.ps1` to run your trusted PowerShell
script after SSH setup, as the local administrator. Set `SOFTWARE_DIR=/path/to/packages`
to copy regular files/directories into `C:\vmctl\software`; reference them with
`Join-Path $PSScriptRoot 'software'`. Neither option installs anything by itself.
See [post-install.ps1.example](post-install.ps1.example) for the extension contract.

The hook runs synchronously in a separate PowerShell process. It must return zero
only on success, check native installer exit codes, and throw or exit nonzero on
failure. Use silent/no-restart installer switches, verify package hashes/signatures,
and log your actions. Do not prompt, download packages in isolated mode, or reboot
inside the hook. Record required reboots and perform them after bootstrap finishes.
There is no automatic retry, resume, rollback or package manager in this hook.
SSH remains available if the optional hook fails, for diagnosis.

## Private outputs and limitations

This example and the Kylin example both default to VNC port 5904. Set a different
`console.vncPort` before creating a second simultaneous instance.

- `generated/password.txt` contains the random Windows password, separate from
  the VNC password. The rendered XML and derived ISO also contain this password.
- `generated/client-key` is the private SSH key and never enters the ISO; only
  `client-key.pub` is copied. Keep the generated directory private. Ports 2225 and
  5904 must be free or changed in the configuration before instance creation.
- Output directories use private permissions. An existing `generated/` causes
  preparation to fail rather than overwrite media or rotate credentials.
  A failed build leaves its partial directory for diagnosis; remove it manually
  only after confirming no instance references it.
- Allow space for the extracted ISO tree and another full ISO (roughly twice
  the input media size), plus the separate VM disk. The original ISO is unchanged.
- Keep the derived ISO available while the instance references it. Never commit
  or publish generated media, rendered answers, passwords or Windows images.
- This is one successful zero-input installation, not a repeatability guarantee
  for every Windows build. A separate offline SSH trial on this build verified
  key login, automatic service startup, guest Secure Boot/TPM readiness and login
  after normal shutdown/start. Activation and RDP remain unverified. The CLI's
  Linux/Kickstart `verify` is not Windows verification; status remains
  `installed-unverified` after `finish-install`.

## References

- [UEFI ISO boot entries](https://learn.microsoft.com/en-us/windows-hardware/manufacture/desktop/oscdimg-command-line-options)
- [Setup product key](https://learn.microsoft.com/en-us/windows-hardware/customize/desktop/unattend/microsoft-windows-setup-userdata-productkey-key)
- [Automating OOBE](https://learn.microsoft.com/en-us/windows-hardware/customize/desktop/automate-oobe)
