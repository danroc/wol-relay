//! Wake-on-LAN relay that listens for WOL packets on specified network
//! interfaces and relays them to other networks.

use std::io::IsTerminal;
use std::net::{Ipv4Addr, SocketAddr, UdpSocket};
use std::process::exit;

use anyhow::{Context, anyhow};

use wol_relay::network::{
    Ipv4Network, MAX_PACKET_SIZE, contains_ip, is_ip_in_any, local_ipv4_addrs, resolve_arg,
    send_wol_packet, sender_socket,
};
use wol_relay::wol;

/// Configures the global logger to write to stdout. If stdout is a TTY, it
/// uses console format; otherwise, it uses JSON format.
fn setup_logger() {
    if std::io::stdout().is_terminal() {
        tracing_subscriber::fmt().init();
    } else {
        tracing_subscriber::fmt().json().init();
    }
}

/// Runs the relay loop. Everything the Go `main` did after logging setup.
fn run() -> anyhow::Result<()> {
    let mut args = std::env::args();
    let argv0 = args.next().unwrap_or_else(|| "wol-relay".to_string());
    let args: Vec<String> = args.collect();
    if args.is_empty() {
        return Err(anyhow!("Usage: {argv0} INTERFACES... [CIDR...]"));
    }

    let mut networks: Vec<Ipv4Network> = Vec::new();
    for arg in &args {
        let resolved = resolve_arg(arg).context("Failed to resolve argument")?;
        networks.extend(resolved);
    }

    if networks.is_empty() {
        return Err(anyhow!("No valid networks found"));
    }

    let self_ips: Vec<Ipv4Addr> =
        local_ipv4_addrs().context("Failed to enumerate local addresses")?;

    let listener = UdpSocket::bind(("0.0.0.0", wol::DEFAULT_PORT))
        .map_err(|err| anyhow!("Cannot start server: {err}"))?;
    let sender = sender_socket().context("Cannot start server")?;

    let mut buffer = [0u8; MAX_PACKET_SIZE];
    tracing::info!("Listening for WOL packets...");

    loop {
        let (n, source) = match listener.recv_from(&mut buffer) {
            Ok(result) => result,
            Err(err) => {
                tracing::error!(error = %err, "Cannot read WOL packet");
                continue;
            }
        };

        // An IPv4 socket only yields IPv4 sources; IPv6 is unreachable.
        let source_ip = match source {
            SocketAddr::V4(addr) => *addr.ip(),
            SocketAddr::V6(_) => continue,
        };

        // Ignore packets from networks that we are not monitoring.
        if !is_ip_in_any(source_ip, &networks) {
            continue;
        }

        // Skip packets from this host to prevent relay loops.
        if contains_ip(&self_ips, source_ip) {
            continue;
        }

        let mac = match wol::parse_packet(&buffer[..n]) {
            Ok(mac) => mac,
            Err(err) => {
                tracing::error!(
                    source_ip = %source_ip,
                    packet_size = n,
                    error = %err,
                    "Failed to parse WOL packet"
                );
                continue;
            }
        };

        for network in &networks {
            // Don't send the WOL packet to the same network as the source IP
            // to avoid broadcasting the packet a second time on the original
            // network.
            if network.contains(source_ip) {
                continue;
            }

            if let Err(err) = send_wol_packet(&sender, *network, mac) {
                tracing::error!(
                    source_ip = %source_ip,
                    target_network = %network,
                    target_mac = wol::mac_to_string(&mac),
                    error = %err,
                    "Failed to relay WOL packet"
                );
            } else {
                tracing::info!(
                    source_ip = %source_ip,
                    target_network = %network,
                    target_mac = wol::mac_to_string(&mac),
                    "WOL packet relayed successfully"
                );
            }
        }
    }
}

fn main() {
    setup_logger();

    if let Err(err) = run() {
        tracing::error!(error = %err, "{err:#}");
        exit(1);
    }
}
