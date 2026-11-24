# Operations and release qualification

## Deployment identity

Run `vmctl version --json` and retain the output with the deployment record.
Record the runtime image identity before starting guests:

```sh
docker image inspect vmctl-runtime:local --format '{{.Id}} {{json .RepoDigests}}'
docker run --rm --network none vmctl-runtime:local qemu-system-x86_64 --version
docker run --rm --network none vmctl-runtime:local \
  dpkg-query -W qemu-system-x86 qemu-utils ovmf swtpm
```

Use a retained runtime image ID or immutable digest through `--image` on every
command for a deployment. Keep the old image until upgrade recovery has passed.
The Dockerfile currently resolves packages at build time; rebuilding it later
does not guarantee the same runtime. Record and retain the image itself.

## Host and Docker restart recovery

Before planned host or Docker maintenance, stop each guest through `vmctl stop`.
Resolve shutdown timeouts before proceeding. Containers do not restart
automatically.

After restart, use the original state directory and runtime image:

```sh
vmctl --state-dir /private/vm-state --image RUNTIME_IMAGE doctor
vmctl --state-dir /private/vm-state --image RUNTIME_IMAGE list
vmctl --state-dir /private/vm-state --image RUNTIME_IMAGE status demo
vmctl --state-dir /private/vm-state --image RUNTIME_IMAGE start demo
```

`start` reuses persistent instance data and recreates a stopped container. After
an unexpected power loss this is a crash boot: guest filesystem and application
recovery still apply. Check guest files and services after login. A paused guest
whose QEMU process was lost cannot resume its former RAM state.

Do not change installation phase merely because the container stopped. An
interrupted installer can be continued with `start` while its phase remains
`installing`; declare completion only after the guest's installation checks pass.

## Interrupted operations

Creation and import clean up incomplete work on handled errors. Abrupt process
termination can leave staging directories; preserve them for diagnosis. If a
name exists but `status` cannot load it, inspect its contents before removing it
or choosing a different name. Never replace its disk with a new blank disk.

Snapshot publication uses a staged copy. Restore replaces multiple files with
runtime-error rollback; host power loss during replacement still requires manual
recovery. Keep the instance stopped and preserve `.restore-*` and `.rollback-*`
directories. Restore from a known-good exported backup into a new name rather
than guessing which disk/NVRAM/TPM files belong together.
`start` refuses instances with these recovery directories so a partially replaced
disk, firmware and TPM set cannot be booted accidentally. After recovery, move
the retained diagnostic directories out of the instance directory before starting.

After checkpoint cancellation or timeout, inspect `status` and QEMU logs before
retrying: command interruption does not cancel QEMU's save/load operation.

## Backup and upgrade drill

1. Record CLI version, runtime image identity, configuration and external media
   and SSH-key locations. Keep their private backups.
2. Stop the instance and export it to storage outside its state directory.
3. Keep the original stopped. Import into a clean state directory, using
   `--config` when media paths or host ports differ.
4. Start with the retained runtime. Check known files, SSH host identity, guest
   services and applicable Secure Boot/TPM status. For an encrypted guest, verify
   its protected storage unlocks as expected.
5. Stop the restored guest. Repeat with the candidate CLI/runtime, retaining the
   old backup and runtime for rollback. Do not run both identities concurrently.

See [bundle details](lifecycle.md#stopped-export-and-import). A backup drill passes
only after guest-side data and identity checks, not merely successful import.

## Monitoring and capacity

Monitor container exit state, `vmctl status`, guest service health and free space
on both instance and backup filesystems. `status` reports container/QEMU state,
not application health. Use `logs` for recent QEMU/container diagnostics.

Budget guest RAM in aggregate plus host/QEMU overhead. Paused guests retain RAM.
Sparse disks can grow toward virtual size; snapshots and exports need additional
copies, and import can expand sparse files. Reserve recovery space before a
backup or restore operation. CPU and disk throughput must be measured against
the actual guest workload.

## Release artifacts

With Go and GNU tar/coreutils available, build a candidate without publishing it:

```sh
VMCTL_COMMIT='<full source commit hash>' make release VERSION=v1.0.0-rc.1
cd dist
sha256sum -c vmctl-v1.0.0-rc.1-linux-amd64.tar.gz.sha256
```

The archive contains the Linux amd64 CLI and license notices. Source and runtime
build instructions are distributed separately through `make dist` and the root
README. Build metadata is supplied by the caller and should identify the exact
reviewed source. The packaging command does not certify release qualification.

Before v1.0, complete [all acceptance gates](v1.0.md), attach test commands and
results to the candidate record, and record runtime identity and known issues.
Publishing a tag, image or release remains an explicit maintainer action.
