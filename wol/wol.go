// Package wol provides functions to build and parse Wake-on-LAN packets.
package wol

import (
	"bytes"
	"errors"
	"net"
)

// Wake-on-LAN constants.
const (
	DefaultPort = 9
	PacketSize  = 102
	HeaderSize  = 6
	MACSize     = 6
	MACRepeat   = 16
)

// Wake-on-LAN errors.
var (
	ErrInvalidMAC    = errors.New("invalid MAC-48 address")
	ErrInvalidSize   = errors.New("invalid Wake-on-LAN packet size")
	ErrInvalidHeader = errors.New("invalid Wake-on-LAN packet header")
)

// Header is the fixed header for Wake-on-LAN packets.
var Header = []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

// ParsePacket parses a Wake-on-LAN packet and returns the MAC address
// contained in it.
func ParsePacket(packet []byte) (net.HardwareAddr, error) {
	if len(packet) != PacketSize {
		return nil, ErrInvalidSize
	}

	buffer := bytes.NewBuffer(packet)
	header := buffer.Next(HeaderSize)

	// Validate the header, it must be the 0xFF byte repeated 6 times.
	for !bytes.Equal(header, Header) {
		return nil, ErrInvalidHeader
	}

	// Extract the MAC address from the packet. It should be the same MAC
	// address repeated 16 times.
	mac := buffer.Next(MACSize)
	for range MACRepeat - 1 {
		if !bytes.Equal(mac, buffer.Next(MACSize)) {
			return nil, ErrInvalidMAC
		}
	}

	return mac, nil
}

// BuildPacket builds a Wake-on-LAN packet with the given MAC address. Only
// MAC-48 addresses are supported.
func BuildPacket(mac net.HardwareAddr) ([]byte, error) {
	if len(mac) != MACSize {
		return nil, ErrInvalidMAC
	}

	packet := make([]byte, 0, PacketSize)
	buffer := bytes.NewBuffer(packet)
	buffer.Write(Header)
	for range MACRepeat {
		buffer.Write(mac)
	}

	return buffer.Bytes(), nil
}
