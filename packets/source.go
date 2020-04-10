package packets

import (
	"github.com/google/gopacket"
)

type PacketSource interface {
	NextPacket() (gopacket.Packet, error)
}
