package packets

import (
	"github.com/google/gopacket"
)

type PacketSink interface {
	NextPacket(gopacket.Packet) error
}

type PacketDataSink interface {
	WritePacketData(buf gopacket.SerializeBuffer) (n int, err error)
}

func NewPacketSink(snk PacketDataSink, opts gopacket.SerializeOptions) PacketSink {
	buf := gopacket.NewSerializeBuffer()

	return &packetSink{
		buf:      buf,
		opts:     opts,
		dataSink: snk,
	}
}

type packetSink struct {
	buf      gopacket.SerializeBuffer
	opts     gopacket.SerializeOptions
	dataSink PacketDataSink
}

func (sink *packetSink) NextPacket(pkt gopacket.Packet) (err error) {
	err = gopacket.SerializePacket(sink.buf, sink.opts, pkt)
	if err != nil {
		return
	}

	_, err = sink.dataSink.WritePacketData(sink.buf)
	return
}
