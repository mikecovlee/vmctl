# Lifecycle

`create CONFIG` allocates private instance metadata. It does not boot a VM.
`install NAME` creates a new disk and boots installation media; it refuses to
overwrite an existing disk. A repeated `start` of a running instance is a no-op.

For manual installation, use VNC and complete the installer. The initial boot
tries installation media; subsequent guest reboots prefer disk. After installation,
`finish-install NAME` records the user's declaration, not a verified result.

For Kickstart, `install NAME --wait` waits for the installer to shut down. A zero
container exit changes the state only to `installed-unverified`. Run `start NAME`
and then `verify NAME`: the current verifier uses the instance SSH key to check
the completion marker and mounted installation media without installing packages.

`status` separates installation phase from current container running state.
`logs` returns recent container logs. `console` gives the VNC endpoint.

`stop NAME` sends ACPI powerdown through QMP and waits up to 60 seconds. A guest
that ignores ACPI causes a nonzero result and is left running. Only `stop NAME
--force` requests a hard power cut. Containers have no automatic restart policy.
Stopping Docker itself does not run this CLI shutdown sequence: stop guests first.

`delete NAME --yes` requires a stopped instance and removes its state and disk.
There is no automatic migration of the historical Kylin shell-script installation.
Snapshot and clone workflows are outside this initial CLI release.
