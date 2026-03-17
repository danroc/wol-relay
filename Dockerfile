# -----------------------------------------------------------------------------
# Builder

FROM golang:1.26.1@sha256:16e774b791968123d6af5ba4eec19cf91c4208cb1f5849efda5d4ffaf6d1c038 AS builder

WORKDIR /app
COPY . .

ARG CGO_ENABLED=0
RUN go build -ldflags="-s -w"

# -----------------------------------------------------------------------------
# Run

FROM scratch

COPY --from=builder /app/wol-relay /app/wol-relay

ENTRYPOINT [ "/app/wol-relay" ]
