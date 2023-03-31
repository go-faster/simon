ARG IMG=gcr.io/distroless/static-debian11
FROM $IMG:nonroot

COPY simon /usr/bin/local/simon

ENTRYPOINT ["/usr/bin/local/simon"]
