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

	frameSrc := PacketSource(br0, layers.LayerTypeEthernet)
	frameSnk := PacketSink(br0, gopacket.SerializeOptions{})
	arpResp := NewArpResponder(br0)

	for {
		frame, err := frameSrc.NextPacket()
		if err == io.EOF {
			log.Fatal("Interface closed!")
		} else if err != nil {
			log.Println("Error:", err)
		}

		log.Println("Got frame:", frame)

		arpLayer := frame.Layer(layers.LayerTypeARP)
		if arpLayer != nil {
			log.Println("ARP message!")

			reply := arpResp.ReplyArp(arpLayer.(*layers.ARP))
			log.Println("Rendered ARP reply:", reply)

			frameSnk.NextPacket(reply)
		}
	}
}