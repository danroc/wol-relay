# -----------------------------------------------------------------------------
# Builder

FROM golang:1.26.3@sha256:db1271a193b5b6cac1749d8d23c3b3564b37852836a5880ca0a65e1924f7feab AS builder

WORKDIR /app
COPY . .

ARG CGO_ENABLED=0
RUN go build -ldflags="-s -w"

# -----------------------------------------------------------------------------
# Run

FROM scratch

COPY --from=builder /app/wol-relay /app/wol-relay

ENTRYPOINT [ "/app/wol-relay" ]
