package main

import (
	"fmt"
	"net"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/danroc/wol-repeater/wol"
)

const (
	MaxPacketSize = 1024
)

func main() {
	// We need at least three arguments: the program name and two interfaces.
	if len(os.Args) < 3 {
		fmt.Printf("Usage: %s INTERFACES...\n", os.Args[0])
		os.Exit(1)
	}

	conn, err := net.ListenUDP("udp", &net.UDPAddr{Port: wol.DefaultPort})
	if err != nil {
		log.Fatalf("Cannot start server: %v\n", err)
	}
	defer conn.Close()

	log.Info("Listening for WOL packets")

	buffer := make([]byte, MaxPacketSize)
	for {
		n, remote, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Errorf("Cannot read WOL packet: %v", err)
		}

		mac, err := wol.ParsePacket(buffer[:n])
		if err != nil {
			log.Warnf("Invalid WOL packet from %s: %v\n", remote, err)
			continue
		}

		log.Infof("Received WOL packet from %s for MAC %s\n", remote, mac.String())
	}
}
