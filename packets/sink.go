package packets

import (
	"github.com/google/gopacket"

	. "github.com/thomas-neuman/spigot/port"
)


type PacketSink interface {
	NextPacket(gopacket.Packet) error
}

type packetSink struct {
	buf			gopacket.SerializeBuffer
	opts		gopacket.SerializeOptions
	dataSink	*packetDataSink
}

func (sink *packetSink) NextPacket(pkt gopacket.Packet) (err error) {
	err = gopacket.SerializePacket(sink.buf, sink.opts, pkt)
	if err != nil {
		return
	}

	_, err = sink.dataSink.WritePacketData(sink.buf)
	return
}

func PacketSinkFromPort(p *Port, opts gopacket.SerializeOptions) *packetSink {
	buf := gopacket.NewSerializeBuffer()
	dataSink := &packetDataSink{
		port: p,
	}

	return &packetSink{
		buf:		buf,
		opts:		opts,
		dataSink:	dataSink,
	}
}


func (dataSink *packetDataSink) WritePacketData(buf gopacket.SerializeBuffer) (n int, err error) {
	return dataSink.port.Write(buf.Bytes())
}

type packetDataSink struct {
	port 	*Port
}