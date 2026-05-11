FROM golang:1.22-alpine AS builder

RUN apk add --no-cache ca-certificates

WORKDIR /build

COPY go.mod go.sum ./
COPY vendor/ vendor/
COPY . .

ARG TARGETOS=linux
ARG TARGETARCH=amd64
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -mod=vendor -ldflags="-s -w" -o iptv-proxy .

FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /build/iptv-proxy /usr/local/bin/iptv-proxy

ENTRYPOINT ["iptv-proxy"]
