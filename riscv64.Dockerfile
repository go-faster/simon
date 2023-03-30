FROM ghcr.io/go-riscv/distroless/static-unstable:nonroot

# riscv64 architecture support for static images
# https://github.com/go-riscv/distroless
# https://github.com/GoogleContainerTools/distroless/issues/1269

COPY simon /usr/bin/local/simon

ENTRYPOINT ["/usr/bin/local/simon"]