package main

import (
	"io"
	"log"
	_ "net"

	_ "github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	_ "github.com/songgao/packets/ethernet"
	_ "github.com/vishvananda/netlink"

	. "github.com/thomas-neuman/spigot/port"
	. "github.com/thomas-neuman/spigot/packets"
	. "github.com/thomas-neuman/spigot/arp"
	"github.com/thomas-neuman/spigot/nkn"
	"github.com/thomas-neuman/spigot/config"
)


func main() {
	config, err := config.GetConfiguration()
	if err != nil {
		log.Fatal(err)
	}

	br0, err := NewPort(config.IfaceName)
	br0.SetUp(config.IPAddress)
	if err != nil {
		log.Fatal(err)
	}

	frameSrc := PacketSourceFromPort(br0, layers.LayerTypeEthernet)

	var procs []PacketProcessor

	arpResp := NewArpResponder(br0)
	procs = append(procs, arpResp)


	nknClient, err := nkn.NewNknClient(config)
	if err != nil {
		log.Fatal(err)
	}
	procs = append(procs, nknClient)

	consumed := false

	for {
		frame, err := frameSrc.NextPacket()
		if err == io.EOF {
			log.Fatal("Interface closed!")
		} else if err != nil {
			log.Println("Error:", err)
		}

		// log.Println("Got frame:", frame)

		for _, proc := range procs {
			frame, consumed = proc.Process(frame)
			if consumed {
				break
			}
			// log.Println("    --> Intermediate packet:", frame)
		}

		if !consumed {
			log.Println("Frame was never consumed!")
		}
	}
}