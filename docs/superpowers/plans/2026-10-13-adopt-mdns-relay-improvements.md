# Adopt mdns-relay Improvements Implementation Plan

> **For agentic workers:** Execute task-by-task under superpowers:executing-plans; steps use checkbox syntax.

**Goal:** Port the engineering improvements from `danroc/mdns-relay` into the Rust wol-relay: cargo-chef Docker build, `.dockerignore`, multi-arch CI with publish smoke test, `--version`, clap CLI with env-var config, flat zerolog-style JSON logs, transient-error classification, graceful SIGTERM/SIGINT shutdown, and skip logging.

**Architecture:** Keep the single-reader relay design (no socket2/SO_BINDTODEVICE redesign — flagged separately). main.rs becomes a supervisor: it parses the CLI, starts the relay in a reader thread plus a signal thread, and waits on a stop channel (mdns-relay's `main.rs` pattern, single-reader variant).

**Spec:** `~/code/github.com/danroc/mdns-relay` at HEAD (bc6aa99) — its `cli.rs`, `logging.rs`, `main.rs`, `relay.rs`, `Cargo.toml`, `Dockerfile`, `.dockerignore`, `.github/workflows/publish-docker.yml` are the reference. Existing wol-relay Go-parity log messages and positional CLI surface (`wol-relay eno1 eno2` / `wol-relay eno1 10.0.0.0/24`) stay working.

**Out of scope (flagged, not applied):** socket2/SO_BINDTODEVICE per-interface architecture; swapping the `scratch` runtime for alpine (our musl binary is proven in scratch).

## Global Constraints

- Backward-compatible CLI: positional targets keep working; new flags are optional (`--version`, `-l/--log-level`, `-F/--log-format`); env vars `WOL_RELAY_TARGETS`, `WOL_RELAY_LOG_LEVEL`, `WOL_RELAY_LOG_FORMAT` (CLI beats env, mdns-relay pattern).
- Log output: flat `{"time":…,"level":"info","message":…,<fields>}`, lowercase level (mdns-relay `logging.rs` port, replaces tracing) — this restores zerolog compatibility.
- Keep existing message texts where the event is unchanged: `Usage: …` via clap help instead, `Listening for WOL packets...`, `Failed to resolve argument`, `No valid networks found`, `Failed to enumerate local addresses`, `Cannot start server`, `Cannot read WOL packet`, `Failed to parse WOL packet`, `Failed to relay WOL packet`, `WOL packet relayed successfully`.
- Log levels (mdns-relay pattern): routine skips → debug, parse failures and send failures → warn, send/relay success and shutdown → info, hard receive error → error then exit nonzero (supervisor restart).
- rustfmt `max_width = 88` (already adopted); conventional commit messages (mdns-relay and renovate style).
- Dockerfile stays `rust:1-alpine` builder → `scratch` runtime, ENTRYPOINT `/app/wol-relay`; adds cargo-chef planner/cook/build stages and `EXPOSE 9/udp`.

## Tasks

1. **CLI module** — `src/cli.rs` (clap derive + env), `Cargo.toml` adds clap; main.rs parses `Cli` but keeps tracing for now. Tests: defaults, multiple targets, flags, env vars + precedence, `--version` (DisplayVersion), no targets → error. Reference: mdns-relay `cli.rs` (adapted: `targets` positional instead of `interfaces`, plus `LogFormat::Auto`).
2. **Logging port** — `src/logging.rs`: mdns-relay `logging.rs` copied + `Auto` format (TTY → text, else JSON; pure `resolve(format, is_tty)` tested). Swap tracing → `log`/`env_logger`/`serde_json` in `Cargo.toml`; convert every log call in main.rs to `log::…!` kv syntax; `RUST_LOG` override. Tests: ported from mdns-relay `logging.rs` tests.
3. **Robust shutdown + error classification** — `network.rs`: `pub fn is_transient(&io::Error) -> bool` + test; `send_wol_packet` returns `io::Result<()>` (test: loopback send Ok). `main.rs`: relay loop moves into a reader thread; signal-hook thread (SIGTERM/SIGINT) → stop channel; error-bridge; `wait_for_stop` (Signal → info `shutting down`, exit 0; ReaderDied → fatal, exit 1). Transient recv → silent continue; hard recv → error + exit. New debug logs for skipped packets; warn for parse/send failures; startup `networks:serde` list. Reference: mdns-relay `main.rs`/`relay.rs`.
4. **Docker + CI** — Dockerfile cargo-chef 3-stage (keep `scratch` runtime, add `EXPOSE 9/udp`); `.dockerignore` (target/, dist/, .git/, docs/, .superpowers/); rewrite `.github/workflows/publish-docker.yml` after mdns-relay's (per-arch native matrix amd64/arm64, temp tags, GHA layer cache, `--version` smoke guard, manifest merge with version + major.minor + major + develop tags, publishes on main pushes and version tags).
5. **Docs + end-to-end verification** — README (options/env vars section, log format note), CHANGELOG Unreleased entries; verify: `make lint test build`, `docker build`, `docker run --rm … --version` smoke, SIGTERM clean shutdown, and the two-bridge relay e2e (packet delivered, one relay log, self-broadcast skipped).
