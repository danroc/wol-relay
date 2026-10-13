# -----------------------------------------------------------------------------
# Builder

FROM rust:1-alpine@sha256:0cce0a5e0e8ba67b455257a3a02a1d99005f382748789d6464460028810f1627 AS builder

WORKDIR /app
COPY . .
RUN cargo build --release

# -----------------------------------------------------------------------------
# Run

FROM scratch

COPY --from=builder /app/target/release/wol-relay /app/wol-relay

ENTRYPOINT [ "/app/wol-relay" ]
