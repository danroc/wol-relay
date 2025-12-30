# -----------------------------------------------------------------------------
# Builder

FROM golang:1.25.5@sha256:6396b3d8039d2050ab7a3c5c6e1cbeed8bf6d2ddc0403e1ab39d78749227ca19 AS builder

WORKDIR /app
COPY . .

ARG CGO_ENABLED=0
RUN go build -ldflags="-s -w"

# -----------------------------------------------------------------------------
# Run

FROM scratch

COPY --from=builder /app/wol-relay /app/wol-relay

ENTRYPOINT [ "/app/wol-relay" ]
