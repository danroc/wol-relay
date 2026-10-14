# syntax=docker/dockerfile:1

# --------------------------------------------------------------------------------------
# Chef — Rust toolchain plus cargo-chef, shared by the planner and build stages.
# --------------------------------------------------------------------------------------

FROM rust:1-alpine@sha256:0cce0a5e0e8ba67b455257a3a02a1d99005f382748789d6464460028810f1627 AS chef

# musl-dev: linker and headers for a static musl target
RUN apk add --no-cache musl-dev \
    && cargo install cargo-chef --locked

WORKDIR /app

# --------------------------------------------------------------------------------------
# Planner — extract a dependency-only recipe from the manifests.
# --------------------------------------------------------------------------------------

FROM chef AS planner

COPY . .
RUN cargo chef prepare --recipe-path recipe.json

# --------------------------------------------------------------------------------------
# Build — compile a static musl binary. Dependencies are built from the recipe first, so
# that layer stays cached until the dependencies change.
# --------------------------------------------------------------------------------------

FROM chef AS build

COPY --from=planner /app/recipe.json recipe.json
RUN cargo chef cook --release --recipe-path recipe.json

COPY . .
RUN cargo build --release

# --------------------------------------------------------------------------------------
# Run — minimal runtime image; the binary is fully static.
# --------------------------------------------------------------------------------------

FROM scratch

COPY --from=build /app/target/release/wol-relay /app/wol-relay

# Wake-on-LAN over UDP; relaying requires host networking (`docker run --net=host`)
EXPOSE 9/udp

ENTRYPOINT [ "/app/wol-relay" ]
