FROM metalstack/builder:latest as builder
RUN make platforms

FROM alpine:3.11
LABEL maintainer="metal-stack Authors <info@metal-stack.io>"
COPY --from=builder /work/bin/metalctl /metalctl
ENTRYPOINT ["/metalctl"]
