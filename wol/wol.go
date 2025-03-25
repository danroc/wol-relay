package wol

import (
	"bytes"
	"net"
)

const (
	DefaultPort = 9
	PacketSize  = 102
	HeaderSize  = 6
	MACSize     = 6
	MACRepeat   = 16
)

// ParsePacket parses a Wake-on-LAN packet and returns the MAC address
// contained in it. If the packet is invalid, it returns nil.
func ParsePacket(packet []byte) net.HardwareAddr {
	if len(packet) != PacketSize {
		return nil
	}

	buffer := bytes.NewBuffer(packet)
	header := buffer.Next(HeaderSize)

	// Validate the header, it must be the 0xFF byte repeated 6 times.
	if len(header) != HeaderSize {
		return nil
	}
	for _, b := range header {
		if b != 0xff {
			return nil
		}
	}

	// Extract the MAC address from the packet. It should be the same MAC
	// address repeated 16 times.
	mac := buffer.Next(MACSize)
	if len(mac) != MACSize {
		return nil
	}
	for range MACRepeat - 1 {
		if !bytes.Equal(mac, buffer.Next(MACSize)) {
			return nil
		}
	}

	return mac
}
