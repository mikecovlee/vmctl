# Release qualification procedure

Use a dedicated Linux x86-64 KVM host and disposable guest state. Keep at most
one guest running on a resource-limited host. Record results against the exact
candidate source and runtime image; retain logs without credentials.

## Automated gate

From an unmodified candidate checkout with Go 1.25 or newer:

```sh
go test -race -shuffle=on -count=1 ./...
go vet ./...
make build
make check
bin/vmctl version --json
```

The default suite skips opt-in KVM tests. A green default run is the source gate,
not a guest qualification result.

## Installation matrix

Follow each final example README in turn: Kylin, Ubuntu, Windows Server and
Windows 11. Prepare from the original ISO into ignored generated media and use a
new empty state directory. For each row record:

| Field | Required evidence |
| --- | --- |
| Source | Candidate commit and CLI version output |
| Runtime | Image ID/digest and package versions from the operations guide |
| Media | Original ISO and generated media SHA-256 |
| Installation | Empty disk creation, installer completion, guest OS version |
| Access | SSH login with generated key; installation completion marker |
| Lifecycle | Normal ACPI shutdown, stopped status, start and SSH reconnection |
| Identity | SSH host key; NVRAM/TPM and Secure Boot checks where applicable |

Windows activation and application support are separate from these checks.
Delete each disposable instance after retaining the evidence.

## Existing opt-in KVM tests

The lifecycle test requires a dedicated running, **verified Kickstart** guest.
It stops that guest and checks that its installation state is retained:

```sh
VMCTL_KVM_TEST=1 \
VMCTL_TEST_STATE=/private/disposable-state \
VMCTL_TEST_INSTANCE=kylin-test \
go test -v ./internal/vm -run '^TestKVMGracefulLifecycle$' -count=1
```

Select network tests only after reading their prerequisites in
`internal/vm/network_kvm_test.go` and `internal/vm/dns_kvm_test.go`. These tests
depend on configured SSH and network endpoints. Prove endpoint reachability
before interpreting a failed guest connection as isolation.

## Recovery matrix

Perform the following on disposable instances only. Before destructive fault
injection, retain a stopped export and external keys on another filesystem.

| Scenario | Pass condition |
| --- | --- |
| Planned host reboot | Stop first; after reboot list/status find the same instance; explicit start and SSH succeed |
| Docker restart | Same discoverability and explicit restart; guest data and identity retained |
| Unexpected QEMU exit | Status reports stopped; explicit start crash-boots the same disk; guest files survive |
| Interrupted create/import | No success response; original instance/backup unchanged; leftovers have a documented cleanup path |
| Interrupted snapshot | Existing snapshots and live files remain usable; incomplete copy is not published as a valid snapshot |
| Interrupted restore | Start blocks retained restore/rollback artifacts; recover coherent data from backup and verify guest login |
| Full instance filesystem | Command fails; original data or documented recovery copy remains available |
| Full backup filesystem | Export fails; existing archive is not overwritten; live instance remains usable |
| Permission failure | Clear nonzero result; no unintended replacement or deletion |
| Port conflict | Start fails before launching a second guest; existing listener remains intact |
| Corrupt bundle | Import rejects it without publishing the requested instance |

Use a size-limited disposable filesystem for disk-full tests, never fill the
host's root filesystem. Do not restart a shared Docker daemon for qualification;
use a dedicated test host. An unexecuted matrix row stays open in the release
checklist.

## Backup, upgrade and soak

Follow the [backup and upgrade drill](operations.md#backup-and-upgrade-drill).
Restore into a clean state root and verify guest files and identities before
removing the original. Run the old and new identity sequentially.

For the 14-day observation gate, record daily host uptime, guest/service health,
free space, runtime exits and relevant logs. Include repeated normal shutdown and
start, a host restart and a successful backup restore. Record any forced power
cut separately. Do not count a paused guest as an application availability test.

Attach dated evidence and unresolved issues to the release record, then update
[v1.0 acceptance](v1.0.md). A skipped scenario is not a passing result.
