//! Wake-on-LAN relay that listens for WOL packets on specified network interfaces and
//! relays them to other networks.
//!
//! Entry point: parses the CLI, sets up logging, spawns the relay reader thread and the
//! signal thread, and waits for shutdown.

use std::net::{Ipv4Addr, SocketAddr, UdpSocket};
use std::process::exit;
use std::sync::mpsc;
use std::thread;

use clap::Parser;

use wol_relay::cli::Cli;
use wol_relay::logging;
use wol_relay::network::{
    self, Ipv4Network, MAX_PACKET_SIZE, is_ip_in_any, local_ipv4_addrs, resolve_arg,
    send_wol_packet, sender_socket,
};
use wol_relay::wol;

/// Why the process is stopping.
enum Stop {
    /// SIGTERM/SIGINT received; the signal number is logged and the process exits
    /// cleanly.
    Signal(i32),
    /// The reader thread hit an unrecoverable error; exit nonzero so the supervisor
    /// restarts the whole instance.
    ReaderDied,
}

/// Log `msg` at error level and exit with status 1.
///
/// Must only be called after the logger is initialized.
fn fatal(msg: &str) -> ! {
    log::error!("{msg}");
    exit(1);
}

/// Logs a fatal error with the given message and exits. The message carries the log
/// message and `error` carries the full error chain.
fn fatal_error(err: impl Into<anyhow::Error>, message: &str) -> ! {
    log::error!(error = format!("{:#}", err.into()); "{message}");
    exit(1)
}

/// Spawn a named thread; a spawn failure is fatal.
fn spawn_thread(name: &str, f: impl FnOnce() + Send + 'static) {
    thread::Builder::new()
        .name(name.to_owned())
        .spawn(f)
        .unwrap_or_else(|err| fatal(&format!("failed to spawn {name} thread: {err}")));
}

/// Forward SIGTERM/SIGINT to the stop channel.
fn spawn_signal_thread(stop_tx: mpsc::Sender<Stop>) {
    let mut signals = signal_hook::iterator::Signals::new([
        signal_hook::consts::SIGTERM,
        signal_hook::consts::SIGINT,
    ])
    .unwrap_or_else(|err| fatal(&format!("failed to install signal handlers: {err}")));

    spawn_thread("signals", move || {
        for sig in signals.forever() {
            let _ = stop_tx.send(Stop::Signal(sig));
        }
    });
}

/// Block until shutdown is requested; never returns normally.
///
/// * `stop_rx` — receives a signal notification or a reader death.
fn wait_for_stop(stop_rx: mpsc::Receiver<Stop>) -> ! {
    match stop_rx.recv() {
        Ok(Stop::Signal(sig)) => {
            log::info!(signal = sig; "shutting down");
            exit(0);
        }

        Ok(Stop::ReaderDied) | Err(_) => {
            fatal("the reader thread failed; exiting");
        }
    }
}

/// Resolve every CLI target to its monitored IPv4 networks, dropping duplicates while
/// preserving order.
///
/// Aborts on the first unknown target.
fn resolve_networks(targets: &[String]) -> Vec<Ipv4Network> {
    let mut networks = Vec::new();
    for target in targets {
        let resolved = resolve_arg(target)
            .unwrap_or_else(|err| fatal_error(err, "Failed to resolve argument"));

        // An interface with several addresses (e.g. loopback aliases) can resolve to
        // the same network more than once; each network only needs one relay.
        for network in resolved {
            if !networks.contains(&network) {
                networks.push(network);
            }
        }
    }

    if networks.is_empty() {
        fatal("No valid networks found");
    }

    networks
}

/// Relays a WOL packet for `mac` to every monitored network other than the source's
/// own, logging the outcome for each.
fn relay_packet(
    sender: &UdpSocket,
    networks: &[Ipv4Network],
    source_ip: Ipv4Addr,
    mac: wol::Mac,
) {
    let target_mac = wol::mac_to_string(&mac);

    for network in networks {
        // Don't send the WOL packet to the same network as the source IP to avoid
        // broadcasting the packet a second time on the original network.
        if network.contains(source_ip) {
            continue;
        }

        match send_wol_packet(sender, *network, mac) {
            Ok(()) => log::info!(
                source_ip:display = source_ip,
                target_network:display = network,
                target_mac = target_mac;
                "WOL packet relayed successfully"
            ),

            Err(err) if network::is_transient(&err) => log::debug!(
                source_ip:display = source_ip,
                target_network:display = network;
                "send queue full, skipping"
            ),

            Err(err) => log::warn!(
                source_ip:display = source_ip,
                target_network:display = network,
                target_mac = target_mac,
                error:display = err;
                "Failed to relay WOL packet"
            ),
        }
    }
}

/// Blocks forever reading WOL datagrams on `listener` and relaying them to every
/// monitored network other than the source's own. On a hard receive error it logs,
/// sends `Stop::ReaderDied` on `stop`, and returns.
///
/// * `listener` — the relay's receive socket (bound to `0.0.0.0`).
/// * `sender` — the relay's send socket (broadcast enabled).
/// * `networks` — monitored networks; sources outside them are skipped.
/// * `self_ips` — this host's addresses; packets from them are skipped.
/// * `stop` — signalled once before returning on a hard receive error.
fn reader_loop(
    listener: UdpSocket,
    sender: UdpSocket,
    networks: Vec<Ipv4Network>,
    self_ips: Vec<Ipv4Addr>,
    stop: mpsc::Sender<Stop>,
) {
    let mut buffer = [0u8; MAX_PACKET_SIZE];

    loop {
        let (n, source) = match listener.recv_from(&mut buffer) {
            Ok(result) => result,

            Err(err) if network::is_transient(&err) => continue,

            Err(err) => {
                log::error!(error:display = err; "Cannot read WOL packet");
                let _ = stop.send(Stop::ReaderDied);
                return;
            }
        };

        // An IPv4 socket only yields IPv4 sources; IPv6 is unreachable.
        let source_ip = match source {
            SocketAddr::V4(addr) => *addr.ip(),
            SocketAddr::V6(_) => {
                log::debug!("skipping packet with IPv6 source address");
                continue;
            }
        };

        // Ignore packets from networks that we are not monitoring.
        if !is_ip_in_any(source_ip, &networks) {
            log::debug!(
                source_ip:display = source_ip;
                "skipping packet from unmonitored network"
            );
            continue;
        }

        // Skip packets from this host to prevent relay loops.
        if self_ips.contains(&source_ip) {
            log::debug!(
                source_ip:display = source_ip;
                "skipping packet from this host"
            );
            continue;
        }

        let mac = match wol::parse_packet(&buffer[..n]) {
            Ok(mac) => mac,
            Err(err) => {
                log::warn!(
                    source_ip:display = source_ip,
                    packet_size = n as u64,
                    error:display = err;
                    "Failed to parse WOL packet"
                );
                continue;
            }
        };

        relay_packet(&sender, &networks, source_ip, mac);
    }
}

fn main() {
    let cli = Cli::parse();

    logging::init(cli.log_level, cli.log_format);

    let networks = resolve_networks(&cli.targets);

    let self_ips: Vec<Ipv4Addr> = local_ipv4_addrs()
        .unwrap_or_else(|err| fatal_error(err, "Failed to enumerate local addresses"));

    let listener = UdpSocket::bind(("0.0.0.0", wol::DEFAULT_PORT))
        .unwrap_or_else(|err| fatal_error(err, "Cannot start server"));

    let sender =
        sender_socket().unwrap_or_else(|err| fatal_error(err, "Cannot start server"));

    let network_names: Vec<String> = networks.iter().map(|n| n.to_string()).collect();
    log::info!(
        networks:serde = network_names;
        "Listening for WOL packets..."
    );

    let (stop_tx, stop_rx) = mpsc::channel::<Stop>();

    spawn_thread("reader", {
        let stop_tx = stop_tx.clone();
        move || reader_loop(listener, sender, networks, self_ips, stop_tx)
    });

    spawn_signal_thread(stop_tx);
    wait_for_stop(stop_rx);
}
