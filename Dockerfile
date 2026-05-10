# -----------------------------------------------------------------------------
# Builder

FROM golang:1.26.3@sha256:2981696eed011d747340d7252620932677929cce7d2d539602f56a8d7e9b660b AS builder

WORKDIR /app
COPY . .

ARG CGO_ENABLED=0
RUN go build -ldflags="-s -w"

# -----------------------------------------------------------------------------
# Run

FROM scratch

COPY --from=builder /app/wol-relay /app/wol-relay

ENTRYPOINT [ "/app/wol-relay" ]
