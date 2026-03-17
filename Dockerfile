# -----------------------------------------------------------------------------
# Builder

FROM golang:1.26.1@sha256:318ba178e04ea7655d4e4b1f3f0e81da0da5ff28a2c48681ff0418fb75f5e189 AS builder

WORKDIR /app
COPY . .

ARG CGO_ENABLED=0
RUN go build -ldflags="-s -w"

# -----------------------------------------------------------------------------
# Run

FROM scratch

COPY --from=builder /app/wol-relay /app/wol-relay

ENTRYPOINT [ "/app/wol-relay" ]
