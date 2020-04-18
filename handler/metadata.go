package handler

import (
	"github.com/google/gopacket"
)

type PacketMetadata struct {
	buf    gopacket.SerializeBuffer
	InPort PortId
	Packet gopacket.Packet
}

func (p *PacketMetadata) FromPacketBytes(packet []byte) error {
	p.buf = gopacket.NewSerializeBuffer()

	return nil
}

func (p *PacketMetadata) ToPacketBytes() (packet []byte, err error) {
	opts := gopacket.SerializeOptions{
		ComputeChecksums: true,
		FixLengths:       true,
	}

	var layers []gopacket.SerializableLayer
	for _, l := range p.Packet.Layers() {
		sl, ok := l.(gopacket.SerializableLayer)
		if !ok {
			return nil, nil
		}
		layers = append(layers, sl)
	}

	gopacket.SerializeLayers(p.buf, opts, layers...)
	return p.buf.Bytes(), nil
}