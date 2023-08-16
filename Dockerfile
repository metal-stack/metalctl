FROM metalstack/builder:latest as platforms
RUN make platforms \
 && strip bin/metalctl-linux-amd64 \
 && cp bin/metalctl-linux-amd64 bin/metalctl

FROM alpine:3.18
LABEL maintainer="metal-stack Authors <info@metal-stack.io>"
COPY --from=platforms /work/bin/metalctl /metalctl
ENTRYPOINT ["/metalctl"]
