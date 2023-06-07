package main

import (
	"fmt"
	"net"

	"github.com/Manthan109/dns/pkg/dns"
)

func main() {
	fmt.Println("Starting DNS sever ...")
	packetConn, err := net.ListenPacket("udp", ":53")
	if err != nil {
		panic(err)
	}
	defer packetConn.Close()

	for {
		buf := make([]byte, 512)
		bytesRead, addr, err := packetConn.ReadFrom(buf)
		if err != nil {
			fmt.Printf("Read error from %s: %s", addr.String(), err)
			continue
		}
		go dns.HandlePacket(packetConn, addr, buf[:bytesRead])
	}
}
