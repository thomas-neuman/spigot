package port

import (
	"log"
	"sync"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/songgao/water"

	"github.com/thomas-neuman/spigot/packets"
)

type Port struct {
	Name  string
	Iface *water.Interface

	wLock sync.Locker
}

func (p *Port) Read() (data []byte, len int, err error) {
	data = make([]byte, 1500)
	len, err = p.Iface.Read(data)

	return
}

func (p *Port) Write(data []byte) (len int, err error) {
	p.wLock.Lock()
	defer p.wLock.Unlock()

	return p.Iface.Write(data)
}

func (p *Port) Ingress() *PortIngress {
	return &PortIngress{
		p: p,
	}
}

type PortIngress struct {
	p *Port
}

// Implement packets.PacketProcessor
func (p *PortIngress) Process(input *layers.IPv4) error {
	eth := layers.Ethernet{
		SrcMAC:       p.p.HardwareAddr(),
		DstMAC:       p.p.HardwareAddr(),
		EthernetType: layers.EthernetTypeIPv4,
	}

	buf := gopacket.NewSerializeBuffer()
	gopacket.SerializeLayers(buf, gopacket.SerializeOptions{
		ComputeChecksums: true,
		FixLengths:       true,
	},
		&eth,
		input,
		gopacket.Payload(input.LayerPayload()))

	pkt := gopacket.NewPacket(buf.Bytes(), layers.LayerTypeEthernet, gopacket.DecodeOptions{})
	p.p.PacketSink(gopacket.SerializeOptions{}).NextPacket(pkt)

	return nil
}

type portPacketDataSink struct {
	p *Port
}

func (snk *portPacketDataSink) WritePacketData(buf gopacket.SerializeBuffer) (n int, err error) {
	return snk.p.Write(buf.Bytes())
}

func (p *Port) PacketSink(opts gopacket.SerializeOptions) packets.PacketSink {
	return packets.NewPacketSink(&portPacketDataSink{
		p: p,
	}, opts)
}

func NewPort(name string) (*Port, error) {
	p := &Port{
		Name:  name,
		wLock: &sync.Mutex{},
	}

	log.Println("Creating TAP...")
	config := water.Config{
		DeviceType: water.TAP,
	}
	config.Name = p.Name

	iface, err := water.New(config)
	if err != nil {
		return nil, err
	}
	p.Iface = iface

	return p, nil
}
