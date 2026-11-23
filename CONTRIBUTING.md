# Contributing

Use Go 1.25 or newer. Legacy root shell scripts are retained for migration reference and are not the new runtime.

```sh
go test -race -shuffle=on -count=1 ./...
go vet ./...
go build ./cmd/vmctl
```

Add failing behavioral regression tests before changing contracts. Tests must be independent of existing VMs, fixed ports, credentials, and checkout paths. KVM tests use disposable instance directories and clean up only their own resources.

Never commit installation media, private keys, disks, generated credentials, internal hostnames, or test-instance state. Label untested guests experimental. Public configuration changes require compatibility notes and malformed-input tests.
