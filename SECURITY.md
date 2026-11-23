# Security boundaries

Docker access is privileged host access. vmctl is a local operator tool, not a multi-tenant security boundary. Use trusted installation media and templates. Kickstart templates execute code inside the guest installer.

Instance directories and credentials are private to their owner. Management listeners default to loopback. A non-loopback bind explicitly exposes that service. Classic VNC does not encrypt display traffic; use an SSH tunnel on untrusted networks.

Isolated networking preserves configured incoming forwards. Exposed services can themselves provide indirect external access. Guest isolation does not protect against an untrusted host administrator.

Report vulnerabilities privately to repository maintainers through the hosting platform when available. Do not include private keys, guest data, or proprietary media in public issues.
