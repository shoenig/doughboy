FROM golang:alpine as builder
WORKDIR /build
ADD . /build
RUN go version && \
    go env && \
    CGO_ENABLED=0 GOOS=linux go build

FROM alpine:latest
MAINTAINER Seth Hoenig <seth.a.hoenig@gmail.com>

WORKDIR /opt
COPY --from=builder /build/doughboy /opt

ENTRYPOINT ["/opt/doughboy"]


## Example Build
#     docker build -t shoenig/doughboy .

## Example launch
#     docker run --rm -v $(pwd)/hack:/hack:ro shoenig/doughboy /hack/classic-responder.hcl

