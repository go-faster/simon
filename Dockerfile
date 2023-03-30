FROM gcr.io/distroless/static-debian11:nonroot

COPY simon /usr/bin/local/simon

ENTRYPOINT ["/usr/bin/local/simon"]