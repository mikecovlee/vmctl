# Kylin V10 example

Supply your own licensed Kylin Server V10 SP3 x86_64 ISO as `kylin.iso` in this directory. Media and installed disks are not distributed.

Build the CLI with `make build`, add `bin/` to `PATH`, and run these commands
from the repository root. For simultaneous examples, choose distinct VNC ports
before `create`; the Windows 11 example also defaults to port 5904.

```sh
vmctl image build runtime
vmctl validate examples/kylin-v10/vm.json
vmctl create examples/kylin-v10/vm.json
vmctl install kylin-demo --wait
vmctl start kylin-demo
vmctl verify kylin-demo
vmctl stop kylin-demo
```

Installation shuts down automatically. `--wait` records `installed-unverified`; successful read-only SSH verification records `verified`. Allow the guest SSH service to start before verifying.

The example uses an instance-specific SSH key, disables password login, and installs an ISO-backed package repository. It disables SELinux for this guest profile, not as a generic runtime requirement. SSH uses localhost port 2224 and VNC port 5904. No host-service forwarding exception is enabled.

`vmctl console kylin-demo --show-secret` displays the generated VNC password. Keep the ISO available for the package repository. `delete --yes` permanently removes the selected stopped instance.
