# -----------------------------------------------------------------------------
# Builder

FROM golang:1.25.5@sha256:97be07314ef2af5f56d22c3bb608c4cffa2a92b3c8252e9f674081ed8217f75b AS builder

WORKDIR /app
COPY . .

ARG CGO_ENABLED=0
RUN go build -ldflags="-s -w"

# -----------------------------------------------------------------------------
# Run

FROM scratch

COPY --from=builder /app/wol-relay /app/wol-relay

ENTRYPOINT [ "/app/wol-relay" ]
