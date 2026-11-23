# Windows 11 assisted installation

Tested with Chinese x64 Windows 11 Pro, build **26300.9457**, image index **4**:
2 vCPUs, 4096 MiB RAM, 64 GiB virtual disk, UEFI/Secure Boot/TPM 2 options,
SATA disk, e1000 NIC, isolated networking. Setup reached the Chinese desktop
and automatically logged into the local `vmctl` administrator account.

This is **assisted**, not fully unattended: the test required a keypress at
the initial DVD boot prompt and selecting **I don't have a product key**.
Partitioning, edition selection, local account creation and offline OOBE
then completed using the answer file. Other builds may behave differently.
Windows activation, guest-side Secure Boot/TPM validation, RDP and Windows
ACPI shutdown acceptance have not been tested. The instance remains
`installed-unverified`; the CLI's Kickstart `verify` is not a Windows verifier.

## Prepare

Supply your own licensed installation media as `windows11.iso` here, or edit
`media.iso` in `vm.json`. Inspect `sources/install.wim` with `wiminfo` and
confirm Pro is index 4 **before** using the template. Adjust `/IMAGE/INDEX`
and locale settings for other media. The answer file **wipes guest disk 0**
and accepts the installation EULA; use it only on a new disposable instance.

Install `openssl` and `genisoimage` on your build machine, then run from the
repository root:

```sh
bash examples/windows11-assisted/prepare.sh
vmctl create examples/windows11-assisted/vm.json
vmctl install windows11-assisted
vmctl console windows11-assisted --show-secret
```

Connect a VNC client to `127.0.0.1:5903` promptly and press a key when asked
to boot the DVD. If missed, use the firmware boot menu to select the DVD.
Select the no-product-key option if prompted. Let Setup reboot into the
disk without pressing a DVD boot key again. Once the desktop is ready:

```sh
vmctl finish-install windows11-assisted
vmctl status windows11-assisted
```

The Windows password is in `generated/password.txt`; it is different from
the VNC password. The template requests one automatic login. Generated XML
and ISO contain the password in plaintext: keep them private and out of
version control. `prepare.sh` refuses an existing `generated/` directory
to avoid silently rotating credentials or replacing media in use.
Keep ISO paths available while the instance configuration references them.
No Windows ISO, product key, generated password or VM disk is distributed.
