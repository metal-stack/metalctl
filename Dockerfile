FROM alpine:3.24
ARG TARGETARCH
ARG TARGETVARIANT
LABEL maintainer="metal-stack authors <info@metal-stack.io>"
COPY --chmod=0755 bin/metalctl-linux-${TARGETARCH}${TARGETVARIANT:+-${TARGETVARIANT}} /metalctl
ENTRYPOINT ["/metalctl"]
