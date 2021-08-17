FROM metalstack/builder:latest as builder
RUN make platforms \
 && strip bin/metalctl-linux-amd64 bin/metalctl

FROM alpine:3.14
LABEL maintainer="metal-stack Authors <info@metal-stack.io>"
COPY --from=builder /work/bin/metalctl /metalctl
ENTRYPOINT ["/metalctl"]
