# Configuration

Configurations are strict JSON with `apiVersion: "v1alpha1"`. Unknown fields,
trailing JSON values and inputs larger than 1 MiB are rejected. Paths in `media`
and `install.template` are relative to the configuration file, not the shell.

Required fields are `apiVersion`, `name`, `resources` (`cpu`, `memoryMiB`, `diskGiB`) and
`media.iso`. Optional `media.extraISOs` attach additional read-only installation
or driver media. The caller supplies all ISO files and their licenses.

Defaults are BIOS, VirtIO disk and NIC, manual installation, isolated networking,
and VNC bound to `127.0.0.1:5901`. Each simultaneous instance needs distinct host
ports. `machine` accepts `firmware` (`bios` or `uefi`), `secureBoot`, `tpm`,
`diskBus` (`virtio` or `sata`) and `nic` (`virtio-net-pci` or `e1000`).

`network.mode` is `isolated` or `egress`. Isolated mode uses QEMU SLIRP
`restrict=on`; explicit TCP `forwards` remain available. Each forward has `host`,
`guest` and optional `bind` (default `127.0.0.1`). No guest-to-host proxy is
configured. IPv6 is disabled in this initial runtime. Egress explicitly permits
SLIRP outbound traffic. A proxy running inside a forwarded service can still
relay traffic; this is not a hostile-code sandbox.

`console.vncBind` and `console.vncPort` control the authenticated VNC listener.
Use an SSH tunnel for remote access: classic VNC authentication does not encrypt
the session. `console NAME --show-secret` explicitly reveals the per-instance
credential; ordinary status output does not.

`install.mode` defaults to `manual`. For `kickstart`, provide `template`,
`kernelPath` and `initrdPath`; the latter two are paths inside the ISO. Template
fields are `Hostname`, `MAC` and `PublicKey`. See `examples/kylin-v10`.

State defaults to `$XDG_STATE_HOME/vmctl` or `~/.local/state/vmctl`. Override it
with `--state-dir`. Each instance has its own disk, configuration, credentials,
QMP socket and (when selected) NVRAM/TPM state. Kickstart installation creates
the SSH identity; SSH verification creates its known-hosts file.
Do not copy a running instance or independently restore only its disk when TPM
or firmware state participates in guest identity.
