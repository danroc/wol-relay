package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/danroc/wol-repeater/wol"
)

const (
	MaxPacketSize = 65536
)

func main() {
	if len(os.Args) < 3 {
		fmt.Printf("Usage: %s INTERFACES...\n", os.Args[0])
		os.Exit(1)
	}

	conn, err := net.ListenUDP("udp", &net.UDPAddr{Port: wol.DefaultPort})
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	buffer := make([]byte, MaxPacketSize)
	for {
		n, remote, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Fatal(err)
		}

		mac := wol.ParsePacket(buffer[:n])
		if mac != nil {
			log.Printf("Received WOL packet from %s for MAC %s\n", remote, mac.String())
		}
	}
}
