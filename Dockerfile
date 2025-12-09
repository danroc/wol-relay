# -----------------------------------------------------------------------------
# Builder

FROM golang:1.25.5@sha256:0ece421d4bb2525b7c0b4cad5791d52be38edf4807582407525ca353a429eccc AS builder

WORKDIR /app
COPY . .

ARG CGO_ENABLED=0
RUN go build -ldflags="-s -w"

# -----------------------------------------------------------------------------
# Run

FROM scratch

COPY --from=builder /app/wol-relay /app/wol-relay

ENTRYPOINT [ "/app/wol-relay" ]
