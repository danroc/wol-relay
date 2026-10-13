# Go → Rust Migration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the Go implementation of wol-relay with a behavior-equivalent Rust implementation (binary, tests, Makefile, Dockerfile), then delete the Go code.

**Architecture:** The program is a single-threaded, blocking UDP loop: resolve CLI args (interface names or IPv4 CIDRs) into monitored networks, listen on `0.0.0.0:9`, validate WOL packets, and relay each packet to the broadcast address of every monitored network other than the source's network. Loop prevention: ignore packets from non-monitored networks and from this host's own IPs. The Rust port keeps the same shape: no async runtime, no CLI framework, raw `std::env::args()` parsing, structured logging that switches between console (TTY) and JSON (non-TTY) formats.

**Tech Stack:** Rust edition 2024 (rustc ≥ 1.85). Crates: `thiserror` 2 (typed errors), `anyhow` 1 (main error handling), `if-addrs` 0.13 (interface enumeration), `tracing` 0.1 + `tracing-subscriber` 0.3 with `fmt` + `json` features (logging). Everything else is `std`. Tooling: `cargo fmt`, `cargo clippy`, Docker `rust:1-alpine` builder → `scratch` runtime.

**Spec:** The spec is the current Go implementation, which is the behavior-parity reference:
- `main.go` — arg resolution, network helpers, UDP relay loop, logging setup
- `wol/wol.go` — WOL packet format (6×`0xFF` + 16× repeated 6-byte MAC = exactly 102 bytes)
- `main_test.go`, `wol/wol_test.go` — test cases that must keep passing after the port
- README usage examples and CLI surface (`wol-relay INTERFACES... [CIDR...]`) must remain valid

## Global Constraints

- Behavior parity with the current Go implementation: same CLI surface (positional args, no flags), same listen port (9), same relay and loop-prevention rules, same log messages and log field names.
- Log field names are frozen (Go `main.go:26-30`): `source_ip`, `target_network`, `target_mac`, `packet_size`, plus `error` on errors. Log *format* parity is semantic (level + message + field names), not byte-identical.
- Exact log message strings (Go `main.go`, keep verbatim): `Usage: <argv0> INTERFACES... [CIDR...]`, `Failed to resolve argument`, `No valid networks found`, `Failed to enumerate local addresses`, `Cannot start server`, `Listening for WOL packets...`, `Cannot read WOL packet`, `Failed to parse WOL packet`, `Failed to relay WOL packet`, `WOL packet relayed successfully`.
- WOL packet rules (Go `wol/wol.go`): exact 102-byte length required to parse; header must be six `0xFF`; the 16 MAC repetitions must be identical; UDP port 9.
- IPv4 only: the listener binds `0.0.0.0` (AF_INET socket — no IPv4-mapped IPv6 sources), IPv6 CIDRs are rejected with an "IPv6 not supported" error.
- Binary name stays `wol-relay`; Docker image stays `ghcr.io/danroc/wol-relay`; ENTRYPOINT stays `/app/wol-relay` in a `scratch` runtime image.
- Dependency whitelist: `thiserror`, `anyhow`, `if-addrs`, `tracing`, `tracing-subscriber` only. No async runtime, no CLI parser crate, no `ipnet`/`cidr` crate (CIDR parsing is hand-written std code, matching Go's semantics).
- No new CI workflows, no new features (YAGNI): the port changes language, not behavior.
- Commit message style: plain imperative, no Conventional-Commit prefixes, matching the repo's history (e.g. "Add unit tests to `wol` package", "Implement WOL relaying").
- Rust is not yet installed on this machine — Task 1 installs the toolchain first.

## Review Focus

The five input classes / failure modes the Go code implies but no Go test covers, each pinned by a test in the owning task:

1. **Oversized UDP datagrams** — a 103-byte datagram must be rejected, not just truncated ones (parse requires exactly 102). Pinned in Task 1: `parse_packet_rejects_oversized_packet`.
2. **CIDRs with host bits set** (`10.0.0.5/24`) — must mask to `10.0.0.0/24`; otherwise "don't relay to the source's own network" silently breaks and self-relay loops become possible. Pinned in Task 2: `parse_cidr_masks_host_bits` and `display_masks_host_bits`.
3. **Non-contiguous interface netmasks** (legacy configs, e.g. `255.0.255.0`) — must be skipped when resolving interface args, never panic or produce a wrong prefix. Pinned in Task 3: `netmask_to_prefix` table includes the non-contiguous `None` case; `resolve_arg_rejects_unknown_interface` pins the unknown-interface path.
4. **Directed broadcast sends failing with `EACCES`** — Linux requires `SO_BROADCAST` to send to a broadcast address; without it every relay send fails silently at runtime while unit tests pass. Pinned in Task 4: `sender_socket_sets_broadcast` asserts the socket flag is enabled.
5. **IPv6 inputs from either direction** — IPv6 CIDR args must be rejected with a distinct message (not a generic parse error), and the listener must never yield IPv4-mapped IPv6 sources. Pinned in Task 2: `parse_cidr` table includes `"::1/128" → IPv6NotSupported`; Task 4 binds `0.0.0.0` (AF_INET) so `recv_from` only produces `SocketAddrV4`.

---

### Task 1: Toolchain setup, Cargo scaffold, and `wol` module

**Files:**
- Create: `Cargo.toml`
- Create: `src/main.rs` (module declarations + stub `main`)
- Create: `src/wol.rs`

**Interfaces:**
- Consumes: nothing (first task).
- Produces (used by Tasks 2–6):
  ```rust
  // src/wol.rs
  pub const DEFAULT_PORT: u16 = 9;
  pub const PACKET_SIZE: usize = 102;
  pub const HEADER_SIZE: usize = 6;
  pub const MAC_SIZE: usize = 6;
  pub const MAC_REPEAT: usize = 16;

  pub type Mac = [u8; MAC_SIZE];
  pub type Packet = [u8; PACKET_SIZE];

  #[derive(Debug, Clone, Copy, PartialEq, Eq, thiserror::Error)]
  pub enum WolError {
      #[error("invalid MAC-48 address")]
      InvalidMac,
      #[error("invalid Wake-on-LAN packet size")]
      InvalidSize,
      #[error("invalid Wake-on-LAN packet header")]
      InvalidHeader,
  }

  pub fn build_packet(mac: Mac) -> Packet;
  pub fn parse_packet(packet: &[u8]) -> Result<Mac, WolError>;

  impl std::fmt::Display for Mac // "00:11:22:33:44:55" — colon-separated lowercase hex
  ```
  Design note: Go's `BuildPacket` could fail on a bad MAC length; in Rust the `Mac = [u8; 6]` type makes that state unrepresentable, so `build_packet` is infallible and `InvalidMac` is only reachable from `parse_packet` (mismatched repetitions).

- [ ] **Step 1: Install the Rust toolchain**

Run: `curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y` then `source "$HOME/.cargo/env"`.
Expected: `rustc --version` prints ≥ 1.85.0.

- [ ] **Step 2: Scaffold the Cargo project**

`Cargo.toml`:

```toml
[package]
name = "wol-relay"
version = "0.1.13"
edition = "2024"

[dependencies]
anyhow = "1"
if-addrs = "0.13"
thiserror = "2"

[profile.release]
strip = true
lto = true
```

`src/main.rs` for now:

```rust
mod network;
mod wol;

fn main() {}
```

Create empty `src/network.rs` (filled in Task 2) and `src/wol.rs` containing only constants, the `Mac`/`Packet` type aliases, and `WolError` from the Interfaces block above.
Expected: `cargo build` succeeds.

- [ ] **Step 3: Write the failing tests (port of `wol/wol_test.go`)**

Append to `src/wol.rs`:

```rust
#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn build_and_parse_roundtrip() {
        let mac: Mac = [0x00, 0x11, 0x22, 0x33, 0x44, 0x55];
        let packet = build_packet(mac);
        assert_eq!(packet.len(), PACKET_SIZE);
        assert_eq!(parse_packet(&packet).unwrap(), mac);
    }

    #[test]
    fn parse_packet_rejects_wrong_size() {
        let short = [0u8; PACKET_SIZE - 1];
        assert_eq!(parse_packet(&short), Err(WolError::InvalidSize));
        let long = vec![0xffu8; PACKET_SIZE + 1];
        assert_eq!(parse_packet(&long), Err(WolError::InvalidSize));
    }

    #[test]
    fn parse_packet_rejects_bad_header() {
        let mut packet = build_packet([0x00, 0x11, 0x22, 0x33, 0x44, 0x55]);
        packet[0] = 0x00;
        assert_eq!(parse_packet(&packet), Err(WolError::InvalidHeader));
    }

    #[test]
    fn parse_packet_rejects_mismatched_repeat() {
        let mut packet = build_packet([0x00, 0x11, 0x22, 0x33, 0x44, 0x55]);
        packet[HEADER_SIZE + MAC_SIZE * 5] = 0x99;
        assert_eq!(parse_packet(&packet), Err(WolError::InvalidMac));
    }

    #[test]
    fn mac_displays_colon_separated_hex() {
        let mac: Mac = [0x00, 0x11, 0x22, 0x33, 0x44, 0x55];
        assert_eq!(mac.to_string(), "00:11:22:33:44:55");
    }
}
```

- [ ] **Step 4: Run tests to verify they fail**

Run: `cargo test`
Expected: FAIL — compile errors, `build_packet`, `parse_packet`, and `Display` not implemented.

- [ ] **Step 5: Implement `build_packet`, `parse_packet`, and `Display for Mac`**

Header is `[0xff; 6]`; packet is header followed by `mac` repeated `MAC_REPEAT` times. Parse: exact length check, header check, then compare each of the remaining 15 repetitions against the first. Display: join the six bytes as two-digit lowercase hex with `:`.

- [ ] **Step 6: Run tests to verify they pass**

Run: `cargo test && cargo clippy --all-targets -- -D warnings`
Expected: 6 tests PASS, no clippy warnings.

- [ ] **Step 7: Commit**

```bash
git add Cargo.toml Cargo.lock src/
git commit -m "Add Rust scaffolding with wol packet module"
```

---

### Task 2: `Ipv4Network` type with CIDR parsing and broadcast computation

**Files:**
- Modify: `src/network.rs`
- Test: `src/network.rs` (`#[cfg(test)]` module)

**Interfaces:**
- Consumes: `std::net::Ipv4Addr`.
- Produces (used by Tasks 3–4):

  ```rust
  // src/network.rs
  #[derive(Debug, Clone, Copy, PartialEq, Eq)]
  pub struct Ipv4Network {
      network: Ipv4Addr, // host bits masked off
      prefix: u8,       // 0..=32
  }

  #[derive(Debug, Clone, PartialEq, Eq, thiserror::Error)]
  pub enum NetworkError {
      #[error("invalid CIDR {input:?}")]
      InvalidCidr { input: String },
      #[error("IPv6 not supported: {input:?}")]
      IPv6NotSupported { input: String },
      #[error("invalid prefix length: {input:?}")]
      InvalidPrefix { input: String },
  }

  impl Ipv4Network {
      pub fn new(addr: Ipv4Addr, prefix: u8) -> Option<Self>; // None if prefix > 32; masks host bits
      pub fn network(&self) -> Ipv4Addr;
      pub fn prefix(&self) -> u8;
      pub fn contains(&self, ip: Ipv4Addr) -> bool;
      pub fn broadcast(&self) -> Ipv4Addr;
  }
  impl std::fmt::Display for Ipv4Network; // "10.0.0.0/24" of the masked network

  pub fn parse_cidr(s: &str) -> Result<Ipv4Network, NetworkError>;
  pub fn netmask_to_prefix(mask: Ipv4Addr) -> Option<u8>;
  ```

  Semantics pinned to Go:
  - `parse_cidr("10.0.0.5/24")` == `parse_cidr("10.0.0.0/24")` (host bits masked, like `net.ParseCIDR`).
  - `net.ParseCIDR` rejects prefixes > 128 and non-numeric junk; we reject prefix > 32 or non-numeric with `InvalidPrefix`/`InvalidCidr` respectively.
  - If the address part contains `:` it is IPv6 → `IPv6NotSupported` (distinct error, matching Go's "IPv6 not supported: %q").

- [ ] **Step 1: Write the failing tests (port of `TestParseCIDR`/`TestToBroadcastIP` plus review-focus cases)**

```rust
#[cfg(test)]
mod tests {
    use super::*;
    use std::net::Ipv4Addr;

    fn net(addr: &str, prefix: u8) -> Ipv4Network {
        Ipv4Network::new(addr.parse().unwrap(), prefix).unwrap()
    }

    #[test]
    fn parse_cidr_valid() {
        assert_eq!(parse_cidr("10.0.0.0/24").unwrap(), net("10.0.0.0", 24));
        assert_eq!(parse_cidr("10.0.0.5/24").unwrap(), net("10.0.0.0", 24));
        assert_eq!(parse_cidr("192.168.2.5/32").unwrap(), net("192.168.2.5", 32));
        assert_eq!(parse_cidr("172.16.0.0/16").unwrap(), net("172.16.0.0", 16));
    }

    #[test]
    fn parse_cidr_masks_host_bits() {
        let parsed = parse_cidr("10.0.0.5/24").unwrap();
        assert_eq!(parsed.network(), "10.0.0.0".parse::<Ipv4Addr>().unwrap());
    }

    #[test]
    fn parse_cidr_invalid() {
        assert!(matches!(parse_cidr("not-a-cidr"), Err(NetworkError::InvalidCidr { .. })));
        assert!(matches!(parse_cidr("::1/128"), Err(NetworkError::IPv6NotSupported { .. })));
        assert!(matches!(parse_cidr("10.0.0.1/129"), Err(NetworkError::InvalidPrefix { .. })));
        assert!(matches!(parse_cidr("10.0.0.1/abc"), Err(NetworkError::InvalidPrefix { .. })));
        assert!(matches!(parse_cidr("10.0.0.1"), Err(NetworkError::InvalidCidr { .. })));
    }

    #[test]
    fn display_masks_host_bits() {
        assert_eq!(parse_cidr("10.0.0.5/24").unwrap().to_string(), "10.0.0.0/24");
    }

    #[test]
    fn broadcast() {
        assert_eq!(net("192.168.1.1", 24).broadcast(), "192.168.1.255".parse().unwrap());
        assert_eq!(net("10.0.0.0", 8).broadcast(), "10.255.255.255".parse().unwrap());
        assert_eq!(net("192.168.2.5", 32).broadcast(), "192.168.2.5".parse().unwrap());
    }

    #[test]
    fn contains() {
        let networks = [net("192.168.1.1", 24), net("10.0.0.1", 24)];
        let ip = |s: &str| s.parse::<Ipv4Addr>().unwrap();
        let cases = [
            ("192.168.1.0", true), ("192.168.1.1", true), ("192.168.1.100", true),
            ("192.168.1.255", true), ("192.168.0.0", false), ("10.0.0.0", true),
            ("10.0.0.1", true), ("10.0.0.100", true), ("10.0.0.255", true),
            ("10.0.1.0", false), ("172.16.0.1", false), ("0.0.0.0", false),
        ];
        for (s, expected) in cases {
            assert_eq!(networks.iter().any(|n| n.contains(ip(s))), expected, "{s}");
        }
    }

    #[test]
    fn netmask_to_prefix_cases() {
        let mask = |s: &str| s.parse::<Ipv4Addr>().unwrap();
        assert_eq!(netmask_to_prefix(mask("255.255.255.0")), Some(24));
        assert_eq!(netmask_to_prefix(mask("255.0.0.0")), Some(8));
        assert_eq!(netmask_to_prefix(mask("0.0.0.0")), Some(0));
        assert_eq!(netmask_to_prefix(mask("255.255.255.255")), Some(32));
        assert_eq!(netmask_to_prefix(mask("255.0.255.0")), None); // non-contiguous
    }
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cargo test`
Expected: FAIL — `Ipv4Network`, `parse_cidr`, `netmask_to_prefix` not defined.

- [ ] **Step 3: Implement `Ipv4Network`, `parse_cidr`, and `netmask_to_prefix` in `src/network.rs`**

- `parse_cidr`: `s.split_once('/')` (`None` → `InvalidCidr`); address part containing `:` → `IPv6NotSupported`; `Ipv4Addr::from_str` failure → `InvalidCidr`; prefix part `u8::from_str` failure or value > 32 → `InvalidPrefix`; then `Ipv4Network::new`.
- Mask computation: convert prefix to a `u32` mask via `u32::MAX << (32 - prefix)` (shift by 0 when prefix is 0 — beware the 32-shift overflow; handle `prefix == 0` explicitly or use `(!0u32).checked_shl(32 - prefix as u32).unwrap_or(!0)`).
- `contains`: `(u32::from(ip) & mask) == (u32::from(network) & mask)` (network is pre-masked, but keep the `&` on both sides for clarity/safety).
- `broadcast`: `Ipv4Addr::from(u32::from(network) | !mask)`.
- `netmask_to_prefix`: count leading one-bits of `u32::from(mask)`; return `Some(count)` only if the remaining bits are all zero, else `None`.

- [ ] **Step 4: Run tests to verify they pass**

Run: `cargo test && cargo clippy --all-targets -- -D warnings`
Expected: all tests PASS, no clippy warnings.

- [ ] **Step 5: Commit**

```bash
git add src/network.rs
git commit -m "Add IPv4 network type with CIDR parsing"
```

---

### Task 3: Interface resolution and host-local address detection

**Files:**
- Modify: `src/network.rs`
- Test: `src/network.rs` (`#[cfg(test)]` module)

**Interfaces:**
- Consumes: `if_addrs::get_if_addrs()` (crate `if-addrs`), `Ipv4Network`/`netmask_to_prefix` from Task 2.
- Produces (used by Task 4):

  ```rust
  // src/network.rs
  pub fn resolve_arg(arg: &str) -> anyhow::Result<Vec<Ipv4Network>>;
  pub fn local_ipv4_addrs() -> anyhow::Result<Vec<Ipv4Addr>>;
  pub fn is_ip_in_any(ip: Ipv4Addr, networks: &[Ipv4Network]) -> bool;
  pub fn contains_ip(ips: &[Ipv4Addr], ip: Ipv4Addr) -> bool;
  ```

  Semantics pinned to Go `main.go`:
  - `resolve_arg`: argument containing `/` is a CIDR → single-element vec. Otherwise it is an interface name: look it up among `get_if_addrs()`; not found → error (`anyhow!("resolve interface {arg:?}: not found")`); map each of the interface's IPv4 addresses (`ip` + `netmask` via `netmask_to_prefix`) to an `Ipv4Network`. Addresses with non-contiguous netmasks are skipped (not an error). Empty result is allowed — main() handles the "no valid networks" case.
  - `local_ipv4_addrs`: all IPv4 addresses across all interfaces (Go `localIPs`). Used for loop prevention.
  - Note for porting `main_test.go`: the IPv4-mapped IPv6 test case (`TestContainsIP` line 29) is moot in Rust — `Ipv4Addr` cannot represent mapped forms, and the listener socket is AF_INET.

- [ ] **Step 1: Write the failing tests (port of `TestContainsIP`, `TestIsIPInAny`, `TestLocalIPs`)**

Add to the `tests` module in `src/network.rs`:

```rust
#[test]
fn contains_ip_cases() {
    let ips: Vec<Ipv4Addr> = ["192.168.1.5", "10.0.0.5"]
        .iter().map(|s| s.parse().unwrap()).collect();
    let ip = |s: &str| s.parse::<Ipv4Addr>().unwrap();
    assert!(contains_ip(&ips, ip("192.168.1.5")));
    assert!(contains_ip(&ips, ip("10.0.0.5")));
    assert!(!contains_ip(&ips, ip("192.168.1.6")));
    assert!(!contains_ip(&ips, ip("10.0.0.4")));
    assert!(!contains_ip(&ips, ip("8.8.8.8")));
}

#[test]
fn is_ip_in_any_cases() {
    let networks = vec![net("192.168.1.1", 24), net("10.0.0.1", 24)];
    let ip = |s: &str| s.parse::<Ipv4Addr>().unwrap();
    let cases = [
        ("192.168.1.0", true), ("192.168.1.1", true), ("192.168.1.100", true),
        ("192.168.1.255", true), ("192.168.0.0", false), ("10.0.0.0", true),
        ("10.0.0.1", true), ("10.0.0.100", true), ("10.0.0.255", true),
        ("10.0.1.0", false), ("172.16.0.1", false), ("0.0.0.0", false),
    ];
    for (s, expected) in cases {
        assert_eq!(is_ip_in_any(ip(s), &networks), expected, "{s}");
    }
}

#[test]
fn local_ipv4_addrs_includes_loopback() {
    let ips = local_ipv4_addrs().unwrap();
    assert!(!ips.is_empty(), "expected at least loopback");
    assert!(ips.contains(&"127.0.0.1".parse().unwrap()), "{ips:?}");
}

#[test]
fn resolve_arg_cidr() {
    let networks = resolve_arg("127.0.0.1/8").unwrap();
    assert_eq!(networks, vec![net("127.0.0.0", 8)]);
}

#[test]
fn resolve_arg_loopback_interface() {
    // Loopback interface name is platform-specific (lo0 on macOS, lo on Linux).
    let ifaces = if_addrs::get_if_addrs().unwrap();
    let lo = ifaces.iter().find(|i| i.is_loopback()).expect("loopback interface");
    let networks = resolve_arg(&lo.name).unwrap();
    assert!(networks.iter().any(|n| n.contains("127.0.0.1".parse().unwrap())), "{networks:?}");
}

#[test]
fn resolve_arg_rejects_unknown_interface() {
    assert!(resolve_arg("no-such-interface-xyz").is_err());
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cargo test`
Expected: FAIL — the four functions not defined.

- [ ] **Step 3: Implement the four functions in `src/network.rs`**

`resolve_arg` with `/` → `parse_cidr` mapped into `anyhow::Result` (`.map_err(anyhow::Error::from)`). Interface path: one pass over `get_if_addrs()` filtered by name; build `Ipv4Network::new(addr.ip, netmask_to_prefix(addr.netmask)?)` — use `filter_map` with `netmask_to_prefix(...).and_then(|p| Ipv4Network::new(addr.ip, p))` so non-contiguous masks are skipped. `local_ipv4_addrs`: `get_if_addrs()` → all `IfAddr::V4` IPs.

- [ ] **Step 4: Run tests to verify they pass**

Run: `cargo test && cargo clippy --all-targets -- -D warnings`
Expected: all tests PASS (Go's `TestLocalIPs`/`TestIsIPInAny`/`TestContainsIP` equivalents included), no clippy warnings.

- [ ] **Step 5: Commit**

```bash
git add src/network.rs Cargo.lock
git commit -m "Add interface resolution and local address detection"
```

---

### Task 4: Logging setup, WOL sender, and the relay main loop

**Files:**
- Modify: `Cargo.toml` (add `tracing`, `tracing-subscriber`)
- Modify: `src/main.rs` (full port of Go `main.go`)
- Modify: `src/network.rs` (add `sender_socket`, `send_wol_packet`)
- Test: `src/network.rs` (sender test); `src/main.rs` has no unit tests, matching Go (the loop is exercised by smoke tests below)

**Interfaces:**
- Consumes: `wol::{build_packet, parse_packet, Display for Mac, DEFAULT_PORT, PACKET_SIZE}`, `network::{resolve_arg, local_ipv4_addrs, is_ip_in_any, contains_ip, Ipv4Network}`.
- Produces:

  ```rust
  // src/network.rs
  pub const MAX_PACKET_SIZE: usize = 1024; // Go MaxPacketSize

  pub fn sender_socket() -> anyhow::Result<std::net::UdpSocket>;
  // Binds "0.0.0.0:0" (OS-assigned ephemeral port) and enables SO_BROADCAST.

  pub fn send_wol_packet(
      socket: &std::net::UdpSocket,
      network: Ipv4Network,
      mac: wol::Mac,
  ) -> anyhow::Result<()>;
  // Sends build_packet(mac) to (network.broadcast(), wol::DEFAULT_PORT).
  ```

  ```rust
  // src/main.rs
  fn setup_logger(); // TTY stdout -> console format; non-TTY -> JSON; global default
  fn run() -> anyhow::Result<()>; // everything Go's main() does after logging setup
  fn main();
  // main(): setup_logger(); on Err(e) from run(): tracing::error!("{e:#}"); exit(1).
  // run() with no CLI args returns the usage error:
  //   anyhow!("Usage: {argv0} INTERFACES... [CIDR...]") where argv0 = env::args().next().
  ```

  Loop semantics (port of Go `main.go:218-264` verbatim):
  - Bind listener `UdpSocket::bind(("0.0.0.0", DEFAULT_PORT))` — error context `Cannot start server`.
  - Buffer `[0u8; MAX_PACKET_SIZE]`; log `Listening for WOL packets...`.
  - `recv_from`: on error log `Cannot read WOL packet` (with `error` field) and continue — never exit.
  - Only `SocketAddrV4` sources are possible (AF_INET socket); extract `Ipv4Addr` via `match`, treat `SocketAddr::V4` as the only arm (V6 unreachable).
  - Skip if `!is_ip_in_any(source_ip, &networks)`.
  - Skip if `contains_ip(&self_ips, source_ip)` (loop prevention).
  - `parse_packet(&buf[..n])`: on error log `Failed to parse WOL packet` with `source_ip`, `packet_size`, `error` fields; continue.
  - For each network not containing the source IP: send via `send_wol_packet`; log `WOL packet relayed successfully` (info) or `Failed to relay WOL packet` (error) with `source_ip`, `target_network`, `target_mac`, plus `error` on failure.
  - Stage errors in `run()` carry Go's fatal messages as anyhow contexts: `Failed to resolve argument`, `No valid networks found` (as a bail!), `Failed to enumerate local addresses`, `Cannot start server`.
  - Sender is one socket created once before the loop (Go opened a socket per send; a single persistent socket with `SO_BROADCAST` is the Rust port's deliberate simplification).

- [ ] **Step 1: Write the failing test pinning `SO_BROADCAST`**

Add to the `tests` module in `src/network.rs`:

```rust
#[test]
fn sender_socket_sets_broadcast() {
    let socket = sender_socket().unwrap();
    assert!(socket.broadcast().unwrap(), "SO_BROADCAST must be enabled");
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cargo test`
Expected: FAIL — `sender_socket` not defined.

- [ ] **Step 3: Implement `sender_socket` and `send_wol_packet` in `src/network.rs`**

`sender_socket`: `UdpSocket::bind("0.0.0.0:0")?` then `set_broadcast(true)?` (with anyhow context). `send_wol_packet`: `socket.send_to(&build_packet(mac), (network.broadcast(), wol::DEFAULT_PORT))?` (context `failed to send WOL packet`).

- [ ] **Step 4: Run test to verify it pass**

Run: `cargo test && cargo clippy --all-targets -- -D warnings`
Expected: PASS.

- [ ] **Step 5: Implement logging setup and the relay loop in `src/main.rs`**

`Cargo.toml` gains:

```toml
tracing = "0.1"
tracing-subscriber = { version = "0.3", features = ["fmt", "json"] }
```

`setup_logger`: `std::io::stdout().is_terminal()` (std `IsTerminal`, no extra crate) chooses a `tracing_subscriber::fmt` layer vs `tracing_subscriber::fmt::format::Json` layer; install as global default with `fmt().init()`-style construction (the two branches build separate `fmt::Subscriber`s, one with `.json()`) — log to stdout either way.

Then port `run()`/`main()` exactly per the Interfaces block, using `tracing::info!`/`tracing::error!` with the frozen field names, e.g.:

```rust
tracing::info!(
    source_ip = %source_ip,
    target_network = %network,
    target_mac = %mac,
    "WOL packet relayed successfully"
);
```

- [ ] **Step 6: Smoke-test the binary manually**

```bash
cargo build
./target/debug/wol-relay; echo "exit=$?"
# Expected: "Usage: ..." (JSON, since piped) printed to stdout, exit=1

timeout 3 ./target/debug/wol-relay lo0 || true   # macOS loopback; use "lo" on Linux
# Expected log line "Listening for WOL packets..." then killed by timeout
```

If the listener fails to bind with permission denied, verify the port (9) is free and note the platform behavior — do not change the port.

- [ ] **Step 7: Run full verification**

Run: `cargo test && cargo fmt --check && cargo clippy --all-targets -- -D warnings`
Expected: all PASS.

- [ ] **Step 8: Commit**

```bash
git add Cargo.toml Cargo.lock src/
git commit -m "Implement WOL relaying in Rust"
```

---

### Task 5: Build tooling migration (Makefile, Dockerfile, .gitignore)

**Files:**
- Modify: `Makefile` (replace Go targets with cargo equivalents; keep the help/color conventions)
- Modify: `Dockerfile` (Go builder → Rust musl builder)
- Modify: `.gitignore` (add Rust build artifacts)
- Unchanged: `.github/workflows/publish-docker.yml` (it only builds the Dockerfile — verify, don't edit)

**Interfaces:**
- Consumes: `make` target structure from the current Makefile (help header, colors, `$(DIST_DIR)`).
- Produces: `make lint` (`cargo fmt --check` + `cargo clippy --all-targets -- -D warnings`), `make lint-fix` (`cargo fmt`), `make test` (`cargo test`), `make build` (release build copied to `dist/wol-relay`), `make run` (`cargo run --`), `make deps-update` (`cargo update`). Go-only targets (`lint-revive`, `lint-sec`, `lint-vet`, `lint-lines*`, `deps-tidy`, `deps-tools`, `test-coverage`, `gocover-cobertura`) are dropped.
- Dockerfile: builder stage `FROM rust:1-alpine@sha256:<digest>` (musl target by default → static binary, runs in `scratch`); runtime stays `FROM scratch` + `COPY --from=builder ... /app/wol-relay` + same ENTRYPOINT. No cargo-chef, no test stage (parity with the Go Dockerfile).

- [ ] **Step 1: Rewrite the Dockerfile**

Pin the digest of `rust:1-alpine` at implementation time:

```bash
docker buildx imagetools inspect rust:1-alpine --format '{{.Manifest.Digest}}'
```

```dockerfile
# -----------------------------------------------------------------------------
# Builder

FROM rust:1-alpine@sha256:<digest-from-step-above> AS builder

WORKDIR /app
COPY . .
RUN cargo build --release

# -----------------------------------------------------------------------------
# Run

FROM scratch

COPY --from=builder /app/target/release/wol-relay /app/wol-relay

ENTRYPOINT [ "/app/wol-relay" ]
```

- [ ] **Step 2: Update `.gitignore`**

Replace `coverage.out`/`coverage.xml` with `/target` (keep `dist/`).

- [ ] **Step 3: Rewrite the Makefile**

Keep: header block, color variables, `$(DIST_DIR)` rule, help target, `.DEFAULT_GOAL := help`. New target set per the Interfaces block. `build` copies `target/release/wol-relay` into `$(DIST_DIR)` (release profile already strips via `Cargo.toml` — parity with Go's `-ldflags "-s -w"`).

- [ ] **Step 4: Verify the Docker build end-to-end**

```bash
docker build -t wol-relay-test .
docker run --rm wol-relay-test; echo "exit=$?"
# Expected: usage message (JSON) to stdout, exit=1 — proves the static binary runs in scratch
```

Expected: build succeeds, container runs and prints the usage message with exit code 1.

- [ ] **Step 5: Verify the Makefile targets**

Run: `make lint test build`
Expected: all three succeed; `dist/wol-relay` exists.

- [ ] **Step 6: Commit**

```bash
git add Makefile Dockerfile .gitignore
git commit -m "Migrate build tooling to Rust"
```

---

### Task 6: Remove Go, update docs, changelog, and Renovate config

**Files:**
- Delete: `go.mod`, `go.sum`, `main.go`, `main_test.go`, `wol/`, `.golangci.yml`, `revive.toml`
- Modify: `README.md` (build/dev instructions, requirements), `CHANGELOG.md` (Unreleased entry), `renovate.json` (drop Go-specific bits)
- Unchanged: `LICENSE`, `.github/workflows/publish-docker.yml`, `renovate.json` base (config:best-practices enables the cargo manager by default)

**Interfaces:**
- Consumes: finished Rust project from Tasks 1–5.
- Produces: repo with no Go remnants; docs describing the Rust build.

- [ ] **Step 1: Delete the Go implementation and Go linter configs**

```bash
git rm -r go.mod go.sum main.go main_test.go wol .golangci.yml revive.toml
```

- [ ] **Step 2: Update `renovate.json`**

Remove `"postUpdateOptions": ["gomodTidy"]` and the `matchManagers: ["gomod"]` package rule. Keep `$schema` and `extends: ["config:best-practices"]` (cargo updates are handled by the default cargo manager).

- [ ] **Step 3: Update `README.md`**

- Standalone section: `go build` → `cargo build --release`, binary path `target/release/wol-relay`.
- Requirements: `Go 1.24+` → `Rust 1.85+ (edition 2024)`.
- Development: `go test ./...` → `cargo test`; keep the Docker build line.

- [ ] **Step 4: Update `CHANGELOG.md`**

Under `## [Unreleased]` add a `### Changed` entry: `Rewrite the service in Rust`. Link style follows existing entries once the commit exists — leave the plain entry without a link if unsure (the maintainer adds links at release time, as the current Unreleased section shows).

- [ ] **Step 5: Verify no Go remnants and full green**

Run: `grep -ri 'golang\|go mod\|go build\|go test' --include='*' . --exclude-dir=.git --exclude-dir=target --exclude-dir=dist || echo clean`
Expected: `clean` (or hits only in `CHANGELOG.md` history, which stays).
Run: `make lint test build`
Expected: all succeed.

- [ ] **Step 6: Commit**

```bash
git add -A
git commit -m "Remove Go implementation and update documentation"
```

---

## Notes for the implementer

- Work in a worktree created via superpowers:using-git-worktrees before starting Task 1 (the repo is otherwise untouched by the migration until the first commit).
- Do not reformat or restructure anything beyond what the tasks list; the port should be the smallest change that keeps behavior identical.
- If a crate version above the pinned major exists at implementation time (e.g. `if-addrs 0.14+`, `thiserror 3+`), prefer the pinned major unless it fails to compile; note any bump in the task's commit.
- `sender_socket` binding `0.0.0.0:0` gives an OS-assigned ephemeral source port, like Go's per-send `DialUDP`; a single reused socket is the accepted deviation.
- Log format is semantic parity only: zerolog emits `{"level":"info",...}` while tracing-subscriber JSON nests fields — field names, level, and message text are what must match.
