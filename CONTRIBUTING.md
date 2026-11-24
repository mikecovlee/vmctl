# Contributing

Use Go 1.25 or newer. Automated coding assistants should also follow [AGENTS.md](AGENTS.md).

```sh
go test -race -shuffle=on -count=1 ./...
go vet ./...
go build ./cmd/vmctl
make check
```

Add failing behavioral regression tests before changing contracts. Tests must be independent of existing VMs, fixed ports, credentials, and checkout paths. KVM tests use disposable instance directories and clean up only their own resources.

`make check` validates shell syntax, example configurations, local Markdown links
and anchors, source-archive contents, and the extracted archive build. It does
not require installation media or KVM.

Never commit installation media, private keys, disks, generated credentials, internal hostnames, or test-instance state. Label untested guests experimental. Public configuration changes require compatibility notes and malformed-input tests.
