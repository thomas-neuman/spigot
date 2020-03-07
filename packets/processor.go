package packets

import (
	"github.com/google/gopacket"
)

type PacketProcessor interface {
	Process(input gopacket.Packet) (output gopacket.Packet, consumed bool)
}