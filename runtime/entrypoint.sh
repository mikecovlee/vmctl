#!/bin/sh
set -eu
umask 077

if [ "${1:-}" != qemu-system-x86_64 ]; then
    exec "$@"
fi

if [ "${VMCTL_TPM:-false}" = true ]; then
    mkdir -p /instance/tpm
    swtpm socket --tpm2 --tpmstate dir=/instance/tpm \
        --ctrl type=unixio,path=/instance/tpm/socket --daemon
fi

exec "$@"
