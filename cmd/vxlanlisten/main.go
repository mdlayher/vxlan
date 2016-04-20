package main

import (
	"log"
	"net"

	"github.com/mdlayher/vxlan"
)

func main() {
	ifi, err := net.InterfaceByName("eth1")
	if err != nil {
		log.Fatal(err)
	}

	c, err := vxlan.NewClient(ifi, net.IPv4(239, 1, 1, 1))
	if err != nil {
		log.Fatal(err)
	}

	for {
		f, addr, err := c.Read()
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("%s: VXLAN ID: %d, EtherType: 0x%04X", addr, f.VNI, uint16(f.Ethernet.EtherType))
	}
}
