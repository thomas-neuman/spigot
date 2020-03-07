package arp

import (
	"log"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	. "github.com/thomas-neuman/spigot/packets"
	. "github.com/thomas-neuman/spigot/port"
)

type ArpResponder struct {
	ethTmpl layers.Ethernet
	arpTmpl layers.ARP
	frameSnk PacketSink
}

func NewArpResponder(p *Port) *ArpResponder {
	ar := &ArpResponder{}

	src := p.HardwareAddr()

	ar.ethTmpl = layers.Ethernet{
		SrcMAC:       net.HardwareAddr(src),
		DstMAC:       []byte{},
		EthernetType: layers.EthernetTypeARP,
	}
	ar.arpTmpl = layers.ARP{
		AddrType:          	layers.LinkTypeEthernet,
		Protocol:          	layers.EthernetTypeIPv4,
		HwAddressSize:     	6,
		ProtAddressSize:   	4,
		Operation:         	layers.ARPReply,
		SourceHwAddress:  	[]byte(src),
		SourceProtAddress: 	[]byte{},
		DstHwAddress:      	[]byte{},
		DstProtAddress:		[]byte{},
	}

	ar.frameSnk = PacketSinkFromPort(p, gopacket.SerializeOptions{})

	return ar
}

func (ar *ArpResponder) replyArp(req *layers.ARP) gopacket.Packet {
	eth := ar.ethTmpl
	arp := ar.arpTmpl

	eth.DstMAC = net.HardwareAddr(req.SourceHwAddress)

	arp.SourceProtAddress = req.DstProtAddress
	arp.DstHwAddress = req.SourceHwAddress
	arp.DstProtAddress = req.SourceProtAddress

	eth.Payload = arp.LayerContents()

	buf := gopacket.NewSerializeBuffer()
	gopacket.SerializeLayers(buf, gopacket.SerializeOptions{
		ComputeChecksums: true,
		FixLengths: true,
	}, &eth, &arp)

	return gopacket.NewPacket(buf.Bytes(), layers.LayerTypeEthernet, gopacket.DecodeOptions{})
}

func (ar *ArpResponder) Process(input gopacket.Packet) (output gopacket.Packet, consumed bool) {
	arpLayer := input.Layer(layers.LayerTypeARP)
	if arpLayer != nil {
		log.Println("ARP message!")

		reply := ar.replyArp(arpLayer.(*layers.ARP))
		log.Println("Rendered ARP reply:", reply)

		ar.frameSnk.NextPacket(reply)

		return reply, true
	}


	return nil, false
}