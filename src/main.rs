//! Wake-on-LAN relay that listens for WOL packets on specified network
//! interfaces and relays them to other networks.

use std::net::{Ipv4Addr, SocketAddr, UdpSocket};
use std::process::exit;

use clap::Parser;

use wol_relay::cli::Cli;
use wol_relay::logging;
use wol_relay::network::{
    Ipv4Network, MAX_PACKET_SIZE, contains_ip, is_ip_in_any, local_ipv4_addrs,
    resolve_arg, send_wol_packet, sender_socket,
};
use wol_relay::wol;

/// Logs a fatal error with the given message and exits. The message carries
/// the log message and `error` carries the full error chain.
fn fatal(err: anyhow::Error, message: &str) -> ! {
    log::error!(error = format!("{err:#}"); "{message}");
    exit(1)
}

/// Runs the relay: parses arguments, resolves monitored networks, and relays
/// WOL packets between them.
fn run() {
    let cli = Cli::parse();

    logging::init(cli.log_level, cli.log_format);

    let mut networks: Vec<Ipv4Network> = Vec::new();
    for arg in &cli.targets {
        match resolve_arg(arg) {
            Ok(resolved) => networks.extend(resolved),
            Err(err) => fatal(err, "Failed to resolve argument"),
        }
    }

    if networks.is_empty() {
        log::error!("No valid networks found");
        exit(1);
    }

    let self_ips: Vec<Ipv4Addr> = match local_ipv4_addrs() {
        Ok(ips) => ips,
        Err(err) => fatal(err, "Failed to enumerate local addresses"),
    };

    let listener = match UdpSocket::bind(("0.0.0.0", wol::DEFAULT_PORT)) {
        Ok(listener) => listener,
        Err(err) => fatal(anyhow::Error::new(err), "Cannot start server"),
    };
    let sender = match sender_socket() {
        Ok(sender) => sender,
        Err(err) => fatal(err, "Cannot start server"),
    };

    let mut buffer = [0u8; MAX_PACKET_SIZE];
    log::info!("Listening for WOL packets...");

    loop {
        let (n, source) = match listener.recv_from(&mut buffer) {
            Ok(result) => result,
            Err(err) => {
                log::error!(error:display = err; "Cannot read WOL packet");
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
                log::error!(
                    source_ip:display = source_ip,
                    packet_size = n as u64,
                    error:display = err;
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
                log::error!(
                    source_ip:display = source_ip,
                    target_network:display = network,
                    target_mac = wol::mac_to_string(&mac),
                    error = format!("{err:#}");
                    "Failed to relay WOL packet"
                );
            } else {
                log::info!(
                    source_ip:display = source_ip,
                    target_network:display = network,
                    target_mac = wol::mac_to_string(&mac);
                    "WOL packet relayed successfully"
                );
            }
        }
    }
}

fn main() {
    run();
}
