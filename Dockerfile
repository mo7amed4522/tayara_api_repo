# Tayara API — production container.
# Two-stage build: compile a static binary, then copy it into a slim alpine runtime.
# The runtime intentionally does NOT bake config.yaml (it holds DB/email/Firebase secrets),
# so mount it at runtime, e.g.:
#   docker run -d -p 8888:8888 \
#     -v $(pwd)/config.yaml:/app/config.yaml \
#     ghcr.io/<org>/tayara_api:latest

FROM golang:1.27-alpine AS builder

WORKDIR /workspace
COPY . .

RUN apk add --no-cache git ca-certificates tzdata \
    && go env -w CGO_ENABLED=0 GOFLAGS=-mod=mod \
    && go mod download \
    && go build -trimpath -ldflags "-s -w" -o /out/server main.go

FROM alpine:3.24

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S tay && adduser -S -G tay tay \
    && mkdir -p /app && chown tay:tay /app

WORKDIR /app

COPY --from=builder /out/server /app/server
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

USER tay

# The API listens on the port read from config (`system.addr`, default 8888 in config.yaml).
EXPOSE 8888

# Config is expected to be mounted at /app/config.yaml; see core/viper.go for -c parsing.
ENTRYPOINT ["./server", "-c", "config.yaml"]