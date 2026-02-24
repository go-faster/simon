FROM alpine

ARG TARGETPLATFORM

RUN apk add --no-cache bash \
	build-base \
	curl \
	cosign \
	docker-cli \
	docker-cli-buildx \
	git \
	gpg \
	mercurial \
	make \
	openssh-client \
	syft \
	tini \
	upx

COPY $TARGETPLATFORM/go-faster-simon*.apk /tmp/
RUN apk add --no-cache --allow-untrusted /tmp/go-faster-simon*.apk

# Set USER environment variable for Go's user.Current() when cgo is not available
ENV USER=fs

ENTRYPOINT ["/usr/bin/simon"]
