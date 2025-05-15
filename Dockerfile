FROM alpine:3.21
LABEL maintainer="metal-stack authors <info@metal-stack.io>"
COPY bin/metalctl-linux-amd64 /metalctl
ENTRYPOINT ["/metalctl"]
