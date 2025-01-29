FROM golang:alpine AS builder

COPY ./ /build
WORKDIR /build

RUN apk update && \
    apk add --no-cache git make && \
    go get -u golang.org/x/lint/golint && \
    BUILD_FLAG="-trimpath" make

FROM alpine:latest

COPY --from=builder /build/bin/alertmanager-syslog /alertmanager-syslog

