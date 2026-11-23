# Windows 11

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
On a Linux host, install `7z`, `openssl`, `genisoimage`, and `wiminfo` (wimlib).
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
offline OOBE and requests one automatic login. After observing the desktop:

```sh
vmctl finish-install windows11
vmctl status windows11
```

`install.mode` remains `manual`: vmctl boots the media; the answer file drives
Windows Setup. `finish-install` is still a user declaration, not automatic
Windows verification. Initial boot is DVD-first, subsequent boots disk-first;
do not repeatedly force DVD boot on an installed disk.

## Private outputs and limitations

This example and the Kylin example both default to VNC port 5904. Set a different
`console.vncPort` before creating a second simultaneous instance.

- `generated/password.txt` contains the random Windows password, separate from
  the VNC password. The rendered XML and derived ISO also contain this password.
- Output directories use private permissions. An existing `generated/` causes
  preparation to fail rather than overwrite media or rotate credentials.
  A failed build leaves its partial directory for diagnosis; remove it manually
  only after confirming no instance references it.
- Allow space for the extracted ISO tree and another full ISO (roughly twice
  the input media size), plus the separate VM disk. The original ISO is unchanged.
- Keep the derived ISO available while the instance references it. Never commit
  or publish generated media, rendered answers, passwords or Windows images.
- This is one successful zero-input installation, not a repeatability guarantee
  for every Windows build. Activation, guest-side Secure Boot/TPM checks, RDP and
  Windows ACPI shutdown acceptance remain untested. No automatic host-visible
  completion handshake is implemented; status remains `installed-unverified`.

## References

- [UEFI ISO boot entries](https://learn.microsoft.com/en-us/windows-hardware/manufacture/desktop/oscdimg-command-line-options)
- [Setup product key](https://learn.microsoft.com/en-us/windows-hardware/customize/desktop/unattend/microsoft-windows-setup-userdata-productkey-key)
- [Automating OOBE](https://learn.microsoft.com/en-us/windows-hardware/customize/desktop/automate-oobe)
