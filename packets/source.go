package packets

import (
	"github.com/google/gopacket"

	. "github.com/thomas-neuman/spigot/port"
)


type PacketSource interface {
	NextPacket() (gopacket.Packet, error)
}

func PacketSourceFromPort(p *Port, decoder gopacket.Decoder) *gopacket.PacketSource {
	return gopacket.NewPacketSource(PacketDataSourceFromPort(p), decoder)
}


type packetDataSource struct {
	port *Port
}

// Implements gopacket.PacketDataSource
func (src *packetDataSource) ReadPacketData() (data []byte, ci gopacket.CaptureInfo, err error) {
	data, n, err := src.port.Read()

	ci.Length = n
	ci.CaptureLength = n
	data = data[:n]

	return
}

func PacketDataSourceFromPort(p *Port) gopacket.PacketDataSource {
	return &packetDataSource{
		port: p,
	}
}