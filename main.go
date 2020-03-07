package main

import (
	"io"
	"log"
	_ "net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	_ "github.com/songgao/packets/ethernet"
	_ "github.com/vishvananda/netlink"

	. "github.com/thomas-neuman/spigot/port"
	. "github.com/thomas-neuman/spigot/packets"
	. "github.com/thomas-neuman/spigot/arp"
)


func main() {
	config, err := GetConfiguration()
	if err != nil {
		log.Fatal(err)
	}

	br0, err := NewPort(config.IfaceName)
	br0.SetUp(config.IPAddress)
	if err != nil {
		log.Fatal(err)
	}

	frameSrc := PacketSourceFromPort(br0, layers.LayerTypeEthernet)
	frameSnk := PacketSinkFromPort(br0, gopacket.SerializeOptions{})

	var procs []PacketProcessor

	arpResp := NewArpResponder(br0)
	procs = append(procs, arpResp)

	var reply gopacket.Packet
	reply = nil
	consumed := false

	for {
		frame, err := frameSrc.NextPacket()
		if err == io.EOF {
			log.Fatal("Interface closed!")
		} else if err != nil {
			log.Println("Error:", err)
		}

		log.Println("Got frame:", frame)

		for _, proc := range procs {
			reply, consumed = proc.Process(frame)
			if reply != nil {
				frameSnk.NextPacket(reply)
				consumed = true
				break
			}
		}

		if !consumed {
			log.Println("Frame was never consumed!")
		}
	}
}