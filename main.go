package main

import (
	"fmt"
	"net"
	"os"

	"github.com/danroc/wol-repeater/wol"
	log "github.com/sirupsen/logrus"
)

const (
	MaxPacketSize = 1024
)

// ToBroadcastIP calculates the broadcast address for a given IPv4 network.
func ToBroadcastIP(network net.IPNet) (net.IP, error) {
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

// CollectNetworks collects all IPv4 network for the given list of network
// interface names.
func CollectNetworks(interfaces []string) ([]net.IPNet, error) {
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

// SendUDP sends a UDP packet to the given IP and port.
func SendUDP(ip net.IP, port int, packet []byte) (int, error) {
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

// SendWOLPacket sends a Wake-on-LAN packet to the given network and MAC
// address.
func SendWOLPacket(network net.IPNet, mac net.HardwareAddr) error {
	broadcastIP, err := ToBroadcastIP(network)
	if err != nil {
		return err
	}

	packet, err := wol.BuildPacket(mac)
	if err != nil {
		return err
	}

	_, err = SendUDP(broadcastIP, wol.DefaultPort, packet)
	return err
}

func IsIPOneOf(ip net.IP, networks []net.IPNet) bool {
	for _, network := range networks {
		if network.IP.Equal(ip) {
			return true
		}
	}
	return false
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s INTERFACES...\n", os.Args[0])
	}

	networks, err := CollectNetworks(os.Args[1:])
	if err != nil || len(networks) == 0 {
		log.Fatalf("No valid network interfaces found: %v\n", err)
	}

	conn, err := net.ListenUDP("udp", &net.UDPAddr{Port: wol.DefaultPort})
	if err != nil {
		log.Fatalf("Cannot start server: %v\n", err)
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
		if IsIPOneOf(remote.IP, networks) {
			continue
		}

		mac, err := wol.ParsePacket(buffer[:n])
		if err != nil {
			log.Warnf("Invalid WOL packet received from %s: %v\n", remote, err)
			continue
		}

		for _, network := range networks {
			if !network.Contains(remote.IP) {
				log.Infof(
					"Sending WOL packet from %s to %s (MAC: %s)\n",
					remote.String(),
					network.String(),
					mac.String(),
				)

				if err := SendWOLPacket(network, mac); err != nil {
					log.Errorf(
						"Failed to send WOL packet from %s to %s (MAC: %s): %v\n",
						remote.String(),
						network.String(),
						mac.String(),
						err,
					)
				}
			}
		}
	}
}
