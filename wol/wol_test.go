package wol

import (
	"bytes"
	"net"
	"testing"
)

func TestBuildPacketAndParsePacket(t *testing.T) {
	mac := net.HardwareAddr{0x00, 0x11, 0x22, 0x33, 0x44, 0x55}

	packet, err := BuildPacket(mac)
	if err != nil {
		t.Fatalf("BuildPacket failed: %v", err)
	}
	if len(packet) != PacketSize {
		t.Errorf("Packet size = %d, want %d", len(packet), PacketSize)
	}

	parsedMAC, err := ParsePacket(packet)
	if err != nil {
		t.Fatalf("ParsePacket failed: %v", err)
	}
	if !bytes.Equal(mac, parsedMAC) {
		t.Errorf("Parsed MAC = %v, want %v", parsedMAC, mac)
	}
}

func TestBuildPacket_InvalidMAC(t *testing.T) {
	// Invalid MAC length
	mac := net.HardwareAddr{0x00, 0x11, 0x22}

	_, err := BuildPacket(mac)
	if err != ErrInvalidMAC {
		t.Errorf("BuildPacket error = %v, want %v", err, ErrInvalidMAC)
	}
}

func TestParsePacket_InvalidPacketSize(t *testing.T) {
	// Make a packet smaller than expected
	packet := make([]byte, PacketSize-1)

	_, err := ParsePacket(packet)
	if err != ErrInvalidSize {
		t.Errorf("ParsePacket error = %v, want %v", err, ErrInvalidSize)
	}
}

func TestParsePacket_InvalidHeader(t *testing.T) {
	mac := net.HardwareAddr{0x00, 0x11, 0x22, 0x33, 0x44, 0x55}
	packet, _ := BuildPacket(mac)

	// Corrupt the header
	packet[0] = 0x00

	_, err := ParsePacket(packet)
	if err != ErrInvalidHeader {
		t.Errorf("ParsePacket error = %v, want %v", err, ErrInvalidHeader)
	}
}

func TestParsePacket_InvalidMACRepeat(t *testing.T) {
	mac := net.HardwareAddr{0x00, 0x11, 0x22, 0x33, 0x44, 0x55}
	packet, _ := BuildPacket(mac)

	// Corrupt one of the repeated MACs
	packet[HeaderSize+MACSize*5] = 0x99

	_, err := ParsePacket(packet)
	if err != ErrInvalidMAC {
		t.Errorf("ParsePacket error = %v, want %v", err, ErrInvalidMAC)
	}
}
