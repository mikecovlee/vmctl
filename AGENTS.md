# Agent guide

These instructions apply to the whole repository. Read the relevant source and
tests before editing; report observed behavior separately from assumptions.

## Project and scope

vmctl is a Go CLI for Linux x86-64 hosts running Docker, QEMU and KVM. Linux and
Windows are **guest** systems. Keep the lifecycle engine OS-neutral; Kylin and
Windows-specific installation choices belong in examples and templates.

Read `README.md`, `docs/configuration.md`, `docs/lifecycle.md`,
`docs/compatibility.md`, and `docs/validation.md` for the current contract and
evidence. Do not expand into clusters, a web UI, migration or Windows-host
support as part of an unrelated change.

## Code map

| Path | Responsibility |
| --- | --- |
| `cmd/vmctl` | Process entry point and cancellation |
| `internal/cli` | Commands, flags and JSON output |
| `internal/config` | Strict configuration parsing and path resolution |
| `internal/state` | Private instance state, atomic persistence and locks |
| `internal/docker` | Docker subprocess boundary and QEMU arguments |
| `internal/qmp` | QMP handshake, commands, responses and deadlines |
| `internal/vm` | Lifecycle orchestration and installation-state transitions |
| `internal/install` | Kickstart rendering and installation digests |
| `runtime` | Container image and runtime entry point |
| `examples` | Guest configurations and installation templates |
| `scripts/dist.sh` | Explicit source-distribution allowlist |

## Behavioral invariants

- Keep `v1alpha1` parsing strict. Configuration changes need compatibility notes
  and malformed-input tests. Resolve media paths relative to the config file.
- Keep instance state and runtime status separate. `finish-install` is a user
  declaration; a container exiting successfully is not OS verification.
- Preserve per-instance disks, credentials, MACs, QMP sockets, NVRAM and TPM
  state. Check ownership before acting on a container. Do not weaken path or
  lock checks to make a test pass.
- `start` is idempotent. `stop` requests ACPI shutdown; a timeout must not kill
  the VM. Forced power-off requires `--force`; deletion requires `--yes` and a
  stopped instance. Preserve the no-automatic-restart policy.
- Default to isolated networking and localhost management listeners. Explicit
  forwards are exceptions, not proof of unrestricted connectivity. Never add
  implicit host-service forwarding or fixed credentials.
- Keep guest verification read-only. Report its actual checks rather than
  describing it as comprehensive OS or security validation.

## Development and validation

Use Go 1.25 or newer. Prefer small typed packages, wrapped errors and
context-bounded operations. Pass subprocess arguments directly, not through
shell interpolation. Do not log passwords or private keys.

For behavioral changes, write a regression test first and exercise both success
and failure paths. Use temporary directories and dynamically allocated ports;
tests must not depend on an existing personal VM or checkout location.

```sh
gofmt -w <changed-go-files>
go test -race -shuffle=on -count=1 ./...
go vet ./...
make build
```

For packaging changes, run `sh -n scripts/dist.sh`, `make dist`, and inspect
`tar -tzf dist/vmctl-source.tar.gz`. The archive must include the intended source
and documentation, without generated media, credentials, state or Git history.
For documentation-only changes, check links and factual consistency; do not
start VMs merely to validate prose.

## Real VM tests

Real KVM tests are opt-in. Read the selected test's prerequisites before setting
`VMCTL_KVM_TEST=1`, `VMCTL_TEST_STATE`, `VMCTL_TEST_INSTANCE`,
`VMCTL_TEST_SSH_PORT`, or `VMCTL_TEST_NETWORK`. Use a dedicated disposable state
directory and instance: lifecycle tests may stop it, and network tests require
the explicitly selected network mode. An environment variable is not permission
to use somebody else's instance.

- Never stop, delete, reinstall or modify unrelated VMs or services.
- Do not run daemon-wide Docker pruning as test cleanup. Track and remove only
  resources created for the task; report any intentionally retained resources.
- Before interpreting a failed connection as isolation, prove the controlled
  endpoint works. Missing tools, authentication errors and boot delays are test
  failures, not evidence of blocked networking.
- Distinguish firmware/TPM smoke tests from guest installation and guest-side
  validation. The assisted Windows example still requires DVD boot input and
  skipping the product-key prompt; do not call it fully unattended or activated.
- Use user-supplied, appropriately licensed OS media. Keep untested guest
  versions experimental and record exactly which version and checks passed.

## Publication and handoff

Use English commit messages. Preserve unrelated user changes. Do not rewrite
history, publish releases, push branches or change remotes without user approval.

Never commit ISOs, disks, private keys, passwords, rendered answer files,
generated answer ISOs, instance state, screenshots containing secrets, or local
infrastructure addresses. Windows answer media contains credentials: only its
template, generator and documentation belong in Git. Keep generated files under
the example's ignored directory. Do not broaden the distribution allowlist to
recursive directory copies.

In the final report, list changes, exact validation performed, skipped checks
and limitations, plus cleanup status. Never promote a successful smoke test to a
claim that an entire operating system or security boundary is verified.
