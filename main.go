// Package main implements a Wake-on-LAN relay that listens for WOL packets on specified
// network interfaces and relays them to other networks.
package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/danroc/wol-relay/wol"
)

const (
	// MaxPacketSize is the size of the buffer used to read WOL packets.
	MaxPacketSize = 1024
)

// Field names for structured logging.
const (
	FieldSourceIP      = "source_ip"
	FieldTargetNetwork = "target_network"
	FieldTargetMAC     = "target_mac"
	FieldPacketSize    = "packet_size"
)

// parseCIDR parses a CIDR string and returns the resulting IPv4 network.
// It rejects IPv6 addresses and invalid CIDR notation.
func parseCIDR(s string) (net.IPNet, error) {
	_, ipnet, err := net.ParseCIDR(s)
	if err != nil {
		return net.IPNet{}, fmt.Errorf("invalid CIDR %q: %w", s, err)
	}
	if ipnet.IP.To4() == nil {
		return net.IPNet{}, fmt.Errorf("IPv6 not supported: %q", s)
	}
	return *ipnet, nil
}

// toBroadcastIP calculates the broadcast address for a given IPv4 network.
func toBroadcastIP(network net.IPNet) (net.IP, error) {
	var (
		ip   = network.IP.To4()
		mask = network.Mask
	)

	if ip == nil || len(mask) != net.IPv4len {
		return nil, fmt.Errorf("invalid IPv4 network: %s", network.String())
	}

	return net.IPv4(
		ip[0]|^mask[0],
		ip[1]|^mask[1],
		ip[2]|^mask[2],
		ip[3]|^mask[3],
	), nil
}

// isIPOneOf checks if the given IP address matches the IP address of any network in the
// provided list.
func isIPOneOf(ip net.IP, networks []net.IPNet) bool {
	for _, network := range networks {
		if network.IP.Equal(ip) {
			return true
		}
	}
	return false
}

// isIPInAny checks if the given IP address is contained in any of the provided
// networks.
func isIPInAny(ip net.IP, networks []net.IPNet) bool {
	for _, network := range networks {
		if network.Contains(ip) {
			return true
		}
	}
	return false
}

// sendWOLPacket sends a Wake-on-LAN packet to the given network and MAC address.
func sendWOLPacket(network net.IPNet, mac net.HardwareAddr) error {
	broadcastIP, err := toBroadcastIP(network)
	if err != nil {
		return err
	}

	packet, err := wol.BuildPacket(mac)
	if err != nil {
		return err
	}

	_, err = sendUDPPacket(broadcastIP, wol.DefaultPort, packet)
	return err
}

// sendUDPPacket sends a UDP packet to the given IP and port.
func sendUDPPacket(ip net.IP, port int, packet []byte) (int, error) {
	conn, err := net.DialUDP("udp4", nil, &net.UDPAddr{
		IP:   ip,
		Port: port,
	})
	if err != nil {
		return 0, err
	}
	defer func() { _ = conn.Close() }()
	return conn.Write(packet)
}

// setupLogger configures the global logger to write to stdout. If stdout is a TTY, it
// uses console format; otherwise, it uses JSON format.
func setupLogger() {
	var output io.Writer = os.Stdout
	if isatty.IsTerminal(os.Stdout.Fd()) {
		output = zerolog.ConsoleWriter{
			Out: os.Stdout,
		}
	}
	log.Logger = log.Output(output)
}

func main() {
	setupLogger()

	if len(os.Args) < 2 {
		log.Fatal().Msgf("Usage: %s INTERFACES... [CIDR...]", os.Args[0])
	}

	var networks []net.IPNet

	for _, arg := range os.Args[1:] {
		if strings.Contains(arg, "/") {
			ipnet, err := parseCIDR(arg)
			if err != nil {
				log.Fatal().Err(err).Msg("Invalid argument")
			}
			networks = append(networks, ipnet)
		} else {
			iface, err := net.InterfaceByName(arg)
			if err != nil {
				log.Fatal().Err(err).Str("interface", arg).Msg("Failed to resolve interface")
			}

			addrs, err := iface.Addrs()
			if err != nil {
				log.Fatal().Err(err).Str("interface", arg).Msg("Failed to get addresses")
			}

			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
					networks = append(networks, *ipnet)
				}
			}
		}
	}

	if len(networks) == 0 {
		log.Fatal().Msg("No valid networks found")
	}

	conn, err := net.ListenUDP("udp4", &net.UDPAddr{Port: wol.DefaultPort})
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot start server")
	}
	defer func() { _ = conn.Close() }()

	buffer := make([]byte, MaxPacketSize)
	log.Info().Msg("Listening for WOL packets...")

	for {
		n, source, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Error().Err(err).Msg("Cannot read WOL packet")
			continue
		}

		// Ignore packets from networks that we are not monitoring.
		if !isIPInAny(source.IP, networks) {
			continue
		}

		// We check if source IP matches one of the interfaces to avoid an infinite loop
		// when sending WOL packets.
		if isIPOneOf(source.IP, networks) {
			continue
		}

		mac, err := wol.ParsePacket(buffer[:n])
		if err != nil {
			log.Error().Err(err).
				Str(FieldSourceIP, source.IP.String()).
				Int(FieldPacketSize, n).
				Msg("Failed to parse WOL packet")
			continue
		}

		for _, network := range networks {
			// Don't send the WOL packet to the same network as the source IP to avoid
			// broadcasting the packet a second time on the original network.
			if network.Contains(source.IP) {
				continue
			}

			// Send the WOL packet and log the result.
			logger := log.With().
				Str(FieldSourceIP, source.IP.String()).
				Str(FieldTargetNetwork, network.String()).
				Str(FieldTargetMAC, mac.String()).
				Logger()

			if err := sendWOLPacket(network, mac); err != nil {
				logger.Error().Err(err).Msg("Failed to relay WOL packet")
			} else {
				logger.Info().Msg("WOL packet relayed successfully")
			}
		}
	}
}
