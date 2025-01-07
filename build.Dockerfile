FROM golang:1.23


WORKDIR /tmp/cache
COPY go.mod go.sum /tmp/cache/
RUN go mod download

COPY . /simon
WORKDIR /simon
RUN go build -o simon ./cmd/simon
RUN mv simon /usr/bin/simon

ENTRYPOINT ["/usr/bin/simon"]
