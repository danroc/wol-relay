# -----------------------------------------------------------------------------
# Builder

FROM golang:1.27.0@sha256:65b6f280bf050ec5af12716857e8ea8439d694dbba8f31ceeb7630670071f2bb AS builder

WORKDIR /app
COPY . .

ARG CGO_ENABLED=0
RUN go build -ldflags="-s -w"

# -----------------------------------------------------------------------------
# Run

FROM scratch

COPY --from=builder /app/wol-relay /app/wol-relay

ENTRYPOINT [ "/app/wol-relay" ]
