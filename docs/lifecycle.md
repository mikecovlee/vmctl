# Lifecycle

`create CONFIG` allocates private instance metadata, a new qcow2 disk and,
when requested, private UEFI NVRAM. It requires the runtime image but does not
boot a VM. An existing instance name is rejected.
`install NAME` boots installation media on that disk and requires the `created`
installation state. A repeated `start` of a running instance is a no-op.

For interactive media, use VNC and complete the installer. The Windows 11 example
also uses `manual` mode, but its media drives Setup through an answer file.
The initial boot
tries installation media; subsequent guest reboots prefer disk. After installation,
`finish-install NAME` records the user's declaration, not a verified result.

The Windows example configures SSH during the first administrator login. Check
`C:\vmctl-install-complete.txt` through the example's generated SSH key before
declaring completion. `C:\vmctl-ssh-ready.txt` confirms SSH setup only: an optional
software hook can still fail afterward. Logs and the failure marker are described
in the example README. The generic `verify` command remains specific to
Kickstart/Linux; SSH availability does not make it a Windows verifier.

For Kickstart, `install NAME --wait` waits for the installer to shut down. A zero
container exit changes the state only to `installed-unverified`. Run `start NAME`
and then `verify NAME`: the current verifier uses the instance SSH key to check
the completion marker and mounted installation media without installing packages.

`status` separates installation phase from current container running state.
`logs` returns recent container logs. `console` gives the VNC endpoint.
Use `status NAME`, `logs NAME`, or `console NAME`; `list` enumerates instances.

`stop NAME` sends ACPI powerdown through QMP and waits up to 60 seconds. A guest
that ignores ACPI causes a nonzero result and is left running. Only `stop NAME
--force` requests a hard power cut. Containers have no automatic restart policy.
Stopping Docker itself does not run this CLI shutdown sequence: stop guests first.

`delete NAME --yes` requires a stopped instance and removes its state and disk.
There is no automatic migration of the historical Kylin shell-script installation.
Deleting an instance also removes its snapshots and internal checkpoints.
Clone workflows are not implemented.

## Stopped export and import

```sh
vmctl stop demo
vmctl export demo /private/demo.tar.gz
vmctl import /private/demo.tar.gz restored-demo
vmctl status restored-demo
vmctl start restored-demo
```

Export requires a stopped instance and refuses to overwrite an existing archive.
The versioned gzip/tar bundle contains the qcow2 disk, configuration, installation
record, management credentials, optional SSH keys, NVRAM and persistent TPM files.
Each payload file has a SHA-256 checksum. Import rejects unsafe paths, links,
duplicate entries, unexpected files and checksum mismatches, and publishes the
new instance directory only after validation. It refuses an existing instance
name and leaves the imported instance stopped.
Import also validates management credentials and runs `qemu-img info/check`
using the runtime image. Only self-contained qcow2 disks without backing files
are accepted. Required NVRAM/TPM components must have the expected types.

The archive contains secrets and is **not encrypted**; store and transfer it as
private material. Checksums detect corruption, not malicious replacement of a
bundle and its manifest. Import only trusted exports. ISO files, runtime sockets,
external snapshot directories and the live QEMU process are not exported. Internal
qcow2 checkpoints remain part of the copied disk and can contain saved guest RAM;
treat the bundle as potentially containing memory secrets as well as disk data.

To relocate media or change host ports, pass a complete configuration:

```sh
vmctl import /private/demo.tar.gz restored-demo --config local.json
```

Only media, network and console settings are taken from this configuration;
machine and resource settings must match the export. Referenced local media must
exist. Import preserves the guest's MAC, credentials and TPM identity: this is
restoration, not identity regeneration or Sysprep cloning. Avoid concurrently
running the original and restored guest on a shared network. Keep the runtime,
CPU and firmware compatible; cross-host compatibility is not automatically
certified. The existing installation phase is preserved, not re-verified.

Preparation requires space for a private staging copy. Tar extraction does not
preserve sparse allocation, so budget for expanded disk files even when the
compressed archive is small. Imports are limited to 1 TiB of declared payload
and 10,000 entries.
Individual JSON/key metadata entries are limited to 1 MiB, the manifest to 4 MiB,
and each NVRAM/TPM file to 64 MiB. Copy loops check cancellation between chunks.

## Pause and resume

`pause NAME` stops guest CPUs through QMP; `resume NAME` continues them. Both are
idempotent for their target state. RAM remains allocated: pausing is not a durable
checkpoint. `status` includes `runtime` (`running`, `paused`, `stopped`, or another
QEMU status); `running` continues to mean the container is alive, including paused
guests. Resume a paused guest before requesting normal ACPI shutdown.

## Stopped snapshots

```sh
vmctl stop demo
vmctl snapshot demo baseline
vmctl snapshots demo
vmctl restore demo baseline --yes
vmctl start demo
```

Snapshots copy the disk, optional UEFI NVRAM, persistent TPM files and installation
record into a private instance-local directory. Running instances are rejected;
existing snapshot labels cannot be overwritten. Restore replaces those files,
requires a stopped instance and explicit confirmation, and leaves it stopped.
Missing required UEFI or TPM state is rejected before any live file is replaced.
These copies do not separately capture live RAM, external installation media, credentials or the
instance configuration. Internal qcow2 checkpoints, including their saved RAM,
remain in the copied disk. Keep the configuration and media compatible. Copying can
consume substantial space. Runtime errors during restore trigger rollback; abrupt
host power loss during its multi-file replacement is not transactionally covered.

## Experimental memory checkpoints

Linux RAM rollback has been tested on Ubuntu 24.04.5 with BIOS, no TPM, 2 vCPUs
and 2 GiB RAM. The same process retained its PID and randomly generated identity
while its memory-only value and counter returned to the saved values. Both CLI
save and restore completed without a timeout in this test. This is not a blanket
compatibility guarantee for all Linux distributions, application consistency or
cross-host restoration. See the [validation record](validation.md#linux-memory-checkpoint).

```sh
vmctl checkpoint demo before-change --yes
vmctl restore-checkpoint demo before-change --yes
vmctl checkpoint demo large-save --yes --timeout 10m
```

These use QEMU internal `savevm`/`loadvm` snapshots inside the qcow2, capturing RAM,
CPU/device state and disk state. They currently require a live QEMU process and
**BIOS without TPM**. UEFI/TPM guests (including the Windows 11 example) are
explicitly rejected: use stopped snapshots instead. Existing checkpoint labels
may be replaced, hence confirmation is required for both operations. Restore can
discard subsequent guest work and can leave the guest running; inspect `status`.
Use the same runtime image, CPU and machine configuration. Cross-version or
cross-host portability and application-level consistency are not guaranteed.
Checkpoint commands wait up to five minutes by default; `--timeout` accepts a
positive duration for both save and restore. Ordinary QMP requests and checkpoint
handshakes retain their ten-second bound. The deadline can expire while QEMU is
still processing a large memory snapshot; a timeout is not proof of cancellation.
An interrupted command reports an unknown outcome. Inspect QEMU before retrying.
