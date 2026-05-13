# -----------------------------------------------------------------------------
# Builder

FROM golang:1.26.3@sha256:633d23bf362cb40dd72b4f277288a8929697d77537f9c801b81aeced19b5bdf3 AS builder

WORKDIR /app
COPY . .

ARG CGO_ENABLED=0
RUN go build -ldflags="-s -w"

# -----------------------------------------------------------------------------
# Run

FROM scratch

COPY --from=builder /app/wol-relay /app/wol-relay

ENTRYPOINT [ "/app/wol-relay" ]
