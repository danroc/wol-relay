# -----------------------------------------------------------------------------
# Builder

FROM golang:1.26.0@sha256:b39810f6440772ab1ddaf193aa0c2a2bbddebf7a877f127c113b103e48fd8139 AS builder

WORKDIR /app
COPY . .

ARG CGO_ENABLED=0
RUN go build -ldflags="-s -w"

# -----------------------------------------------------------------------------
# Run

FROM scratch

COPY --from=builder /app/wol-relay /app/wol-relay

ENTRYPOINT [ "/app/wol-relay" ]
