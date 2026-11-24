# Windows Server 2025

Recommended Windows profile for service deployment and memory-checkpoint experiments:
BIOS/MBR, no TPM, 2 vCPUs, 4 GiB RAM and a 64 GiB disk. Setup runs unattended and
configures key-only SSH automatically. Windows 11 remains available for desktop use.

## Install

Build vmctl and its runtime following the root README; put `bin/` on `PATH`.
Run from the repository root. Host prerequisites: `7z`, `openssl`, `genisoimage`,
`ssh-keygen`, `sha256sum`, and `curl` for automatic OpenSSH downloads.

```sh
bash examples/windows-server2025/prepare.sh /path/to/server2025.iso
vmctl create examples/windows-server2025/vm.json
vmctl install windows-server2025
```

Default image index is **2 (Standard with Desktop Experience)** on the tested ISO.
Check your ISO with `wiminfo`; set `WINDOWS_IMAGE_INDEX` for another index. The
answer template uses zh-CN. For media requiring a setup key, supply
`WINDOWS_PRODUCT_KEY`; installation does not imply activation. Evaluation media
retains Microsoft's activation and expiry limits; non-evaluation media was not
part of this test.

The generator shares OpenSSH bootstrap with the Windows 11 example. It accepts
the same `OPENSSH_MSI`/`OPENSSH_SHA256` local-package overrides, or downloads the
pinned Preview MSI. `POST_INSTALL_SCRIPT` and `SOFTWARE_DIR` support optional
offline software. See [Windows bootstrap details](../windows11/README.md).
Keep both example directories when copying this example.

SSH is localhost **2255**, VNC localhost **5955**. Change ports before creating an
instance if they are occupied. The Windows username is `vmctl`:

```sh
ssh -p 2255 -i examples/windows-server2025/generated/client-key \
  -o IdentitiesOnly=yes -o StrictHostKeyChecking=accept-new \
  -o UserKnownHostsFile=examples/windows-server2025/generated/known_hosts \
  vmctl@127.0.0.1 'Get-Content C:\vmctl-install-complete.txt'
vmctl finish-install windows-server2025
```

Private keys, passwords, rendered answers and the generated ISO stay in ignored
`generated/`. Never publish them. Preparation refuses existing output directories.
The ISO remains required by the instance; do not delete it while referenced.

## Checkpoints

```sh
vmctl checkpoint windows-server2025 before-change --yes
vmctl restore-checkpoint windows-server2025 before-change --yes
```

Tested on Server 2025 Standard Evaluation 26100: zero-input install, SSH, MBR disk,
TPM absent, and restoration of a process's RAM-only value and counter with the
same PID, startup time and random identity. Memory snapshot size was 3.87 GiB.
This is not a cross-host or database-consistency guarantee.

**Known limitation:** a large checkpoint can exceed the current CLI QMP timeout
while QEMU continues saving. The RAM rollback test used a longer-waiting QMP
connection for restoration. Do not assume a timeout cancels the operation or
immediately repeat it. See [lifecycle](../../docs/lifecycle.md).
