FROM alpine:3.18
LABEL maintainer="metal-stack authors <info@metal-stack.io>"
COPY result/metalctl-linux-amd64 /metalctl
ENTRYPOINT ["/metalctl"]
