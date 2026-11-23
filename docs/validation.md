# Validation record

This is a foundation release candidate, not a claim that every guest OS is
certified. Tests ran on a Linux Docker/KVM host using the supplied Kylin V10
x86-64 installation media. Existing legacy VM storage was not modified.

## Executed

- `go test -race -shuffle=on ./...` and `go vet ./...` passed.
- A dedicated Kylin instance completed unattended installation, disk boot and
  SSH acceptance. Guest identity was Kylin Linux Advanced Server V10 (Halberd),
  memory total was 3932 MiB, the only enabled repository was `vmctl-media`, and
  `/dev/sr0` was mounted read-only at `/mnt/iso`.
- QMP ACPI shutdown completed with guest unmount and power-off messages.
  The opt-in `TestKVMGracefulLifecycle` passed. Repeated start was idempotent.
- Two distinct instances ran concurrently. Test instances were subsequently
  stopped and deleted through the CLI.
- Controlled host TCP, UDP and DNS endpoints were reachable in egress mode
  and blocked in isolated mode. Host-side control probes succeeded first.
  The guest had no IPv6 default route; QEMU disables guest IPv6 networking.
- A UEFI/Secure Boot/TPM configuration launched successfully with SATA/e1000.
  This is a firmware startup smoke test, not a Windows installation test.
- Protocol error, QMP timeout, occupied port, state corruption, cancellation,
  disk creation rollback and stop-without-forced-kill tests passed.
- The source archive was inspected: legacy scripts, Git history, credentials,
  ISO images and VM disks were absent.

## Limits of this evidence

- No Windows ISO was installed. Windows examples remain experimental.
- The DNS test exercises a controlled UDP DNS endpoint, not every behavior of
  SLIRP's built-in DNS resolver. IPv6 has a route/configuration check rather
  than an enabled-IPv6 egress comparison.
- Stop timeout preservation is tested with a protocol peer and runtime fake;
  a real unresponsive operating system was not deliberately induced.
- `verify` checks the installation marker and media mount using read-only SSH.
  OS, memory and repository checks above were separate acceptance commands.
- Independent review was not run: execution was performed directly, as
  requested. Source was published to https://github.com/mikecovlee/vmctl
  after user authorization. No release artifact or version tag was published.

Run ordinary tests with `make test`. KVM integration tests require explicit
`VMCTL_KVM_TEST=1`, `VMCTL_TEST_STATE` and `VMCTL_TEST_INSTANCE` settings;
network tests additionally require `VMCTL_TEST_SSH_PORT`. Use disposable,
dedicated instances, never production data.
