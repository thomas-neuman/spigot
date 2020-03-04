package arp

import (
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	. "github.com/thomas-neuman/spigot/port"
)

type ArpResponder struct {
	ethTmpl layers.Ethernet
	arpTmpl layers.ARP
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

	return ar
}

func (ar *ArpResponder) ReplyArp(req *layers.ARP) gopacket.Packet {
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