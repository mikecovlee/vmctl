FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends \
    qemu-system-x86 qemu-utils ovmf swtpm p7zip-full cpio openssh-client ca-certificates \
    && rm -rf /var/lib/apt/lists/*
COPY entrypoint.sh /usr/local/bin/vmctl-runtime
RUN chmod 755 /usr/local/bin/vmctl-runtime
WORKDIR /instance
ENTRYPOINT ["/usr/local/bin/vmctl-runtime"]
