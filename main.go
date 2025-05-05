package main

import (
	"fmt"
	"net"
	"os"

	"github.com/danroc/wol-relay/wol"
	log "github.com/sirupsen/logrus"
)

const (
	MaxPacketSize = 1024
)

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

// collectNetworks collects all IPv4 network for the given list of network
// interface names.
func collectNetworks(interfaces []string) ([]net.IPNet, error) {
	var networks []net.IPNet
	for _, name := range interfaces {
		iface, err := net.InterfaceByName(name)
		if err != nil {
			return nil, err
		}

		addrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok {
				if ipnet.IP.To4() != nil {
					networks = append(networks, *ipnet)
				}
			}
		}
	}
	return networks, nil
}

// sendUDPPacket sends a UDP packet to the given IP and port.
func sendUDPPacket(ip net.IP, port int, packet []byte) (int, error) {
	conn, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   ip,
		Port: port,
	})
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	return conn.Write(packet)
}

// sendWOLPacket sends a Wake-on-LAN packet to the given network and MAC
// address.
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

func isIPOneOf(ip net.IP, networks []net.IPNet) bool {
	for _, network := range networks {
		if network.IP.Equal(ip) {
			return true
		}
	}
	return false
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s INTERFACES...", os.Args[0])
	}

	networks, err := collectNetworks(os.Args[1:])
	if err != nil || len(networks) == 0 {
		log.Fatalf("No valid network interfaces found: %v", err)
	}

	conn, err := net.ListenUDP("udp", &net.UDPAddr{Port: wol.DefaultPort})
	if err != nil {
		log.Fatalf("Cannot start server: %v", err)
	}
	defer conn.Close()

	buffer := make([]byte, MaxPacketSize)
	log.Info("Listening for WOL packets...")

	for {
		n, remote, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Errorf("Cannot read WOL packet: %v", err)
		}

		// We check if remote IP matches one of the interfaces to avoid
		// infinite loop when sending WOL packets.
		if isIPOneOf(remote.IP, networks) {
			continue
		}

		mac, err := wol.ParsePacket(buffer[:n])
		if err != nil {
			log.Warnf("Invalid WOL packet received from %s: %v", remote, err)
			continue
		}

		for _, network := range networks {
			if !network.Contains(remote.IP) {
				if err := sendWOLPacket(network, mac); err != nil {
					log.Errorf(
						"Failed to send WOL packet from %s to %s (MAC: %s): %v",
						remote.IP, network, mac, err,
					)
				} else {
					log.Infof(
						"Sent WOL packet from %s to %s (MAC: %s)",
						remote.IP, network, mac,
					)
				}
			}
		}
	}
}
