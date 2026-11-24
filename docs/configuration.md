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

`network.mode` is `isolated` (default) or `nat`. The legacy `egress` spelling is
accepted and normalized to `nat` when parsing configuration. Isolated mode uses QEMU SLIRP
`restrict=on`; explicit TCP `forwards` remain available. Each forward has `host`,
`guest` and optional `bind` (default `127.0.0.1`). No guest-to-host proxy is
configured. IPv6 is disabled in both modes. NAT explicitly permits
SLIRP outbound traffic. A proxy running inside a forwarded service can still
relay traffic; this is not a hostile-code sandbox.

NAT uses QEMU user-mode networking, not a host bridge or TAP interface. It allows
guest-initiated connections to reachable Internet, LAN and host services; it does
not automatically expose guest services or connect VMs to a shared LAN.
**Forwarding supports TCP only; UDP forwarding is not implemented.** NAT can
carry outbound UDP, which is distinct from an incoming UDP forwarding rule.

Manage existing instances with:

```sh
vmctl network show demo
vmctl stop demo
vmctl network set demo --mode nat
vmctl network forward add demo --bind 127.0.0.1 --host 18080 --guest 8080
vmctl network forward list demo
vmctl network forward remove demo --bind 127.0.0.1 --host 18080
vmctl network set demo --mode isolated
vmctl start demo
```

All mutations require a stopped instance (paused is still running). Updates use
the instance lock and atomic configuration writes, and take effect at the next
start. Editing an original example JSON does not update an existing instance.
Mode changes preserve forwarding rules and guest identity. `status.networkMode`
reports the configured mode. Multiple TCP forwards are supported in either mode;
host ports must be unique within an instance, including its VNC port. Adding a
rule checks current listener availability; startup checks again because other
processes can claim ports afterward. Bind `0.0.0.0` only when LAN exposure is
intended; host and guest firewalls still apply. The guest service must listen.

Forward `bind` and `console.vncBind` accept IPv4 addresses only; IPv6 bindings
are rejected during configuration parsing.

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
