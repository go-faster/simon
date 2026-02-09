FROM ubuntu

RUN apt-get update && apt-get install -y curl

COPY simon /usr/bin/local/simon

ENTRYPOINT ["/usr/bin/local/simon"]
