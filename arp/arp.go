package arp

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type ArpResponder struct {
	replies chan *layers.ARP
}

func NewArpResponder() *ArpResponder {
	return &ArpResponder{}
}

func (a *ArpResponder) DoInit() error {
	a.replies = make(chan *layers.ARP)
	return nil
}

func (a *ArpResponder) FirstLayerType() gopacket.LayerType {
	return layers.LayerTypeARP
}

func (a *ArpResponder) DoWrite(data []gopacket.SerializableLayer) (err error) {
	req := data[0].(*layers.ARP)

	reply := &layers.ARP{
		AddrType:          layers.LinkTypeEthernet,
		Protocol:          layers.EthernetTypeIPv4,
		HwAddressSize:     6,
		ProtAddressSize:   4,
		Operation:         layers.ARPReply,
		SourceHwAddress:   req.DstHwAddress,
		SourceProtAddress: req.DstProtAddress,
		DstHwAddress:      req.SourceHwAddress,
		DstProtAddress:    req.SourceProtAddress,
	}

	a.replies <- reply
	return nil
}

func (a *ArpResponder) DoRead() (data []gopacket.SerializableLayer, err error) {
	reply := <-a.replies
	eth := &layers.Ethernet{
		EthernetType: layers.EthernetTypeARP,
	}

	return []gopacket.SerializableLayer{eth, reply}, nil
}
