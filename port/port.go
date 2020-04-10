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
func (p *PortIngress) Process(input gopacket.Packet) (output gopacket.Packet, consumed bool) {
	output = input
	consumed = false

	/*
		// eth := input.Layer(layers.LayerTypeEthernet)
		el := input.Layer(layers.LayerTypeEthernet).(*layers.Ethernet)
		el.DstMAC = p.p.HardwareAddr()
		// el.SrcMAC = p.p.HardwareAddr()
		el.Length = uint16(len(el.LayerContents()) + len(el.LayerPayload()))
	*/

	eth := layers.Ethernet{
		SrcMAC:       p.p.HardwareAddr(),
		DstMAC:       p.p.HardwareAddr(),
		EthernetType: layers.EthernetTypeIPv4,
	}

	ip4 := gopacket.NewPacket(input.Layer(layers.LayerTypeEthernet).LayerPayload(), layers.LayerTypeIPv4, gopacket.DecodeOptions{})
	// eth.Payload = ip4.Data()

	// ip4 := input.Layer(layers.LayerTypeIPv4).(*layers.IPv4)
	if ip4 == nil {
		return
	}

	// ip4.Payload = ip4.LayerPayload()
	var ls []gopacket.SerializableLayer
	ls = append(ls, &eth)
	for _, l := range ip4.Layers() {
		ls = append(ls, l.(gopacket.SerializableLayer))
	}

	buf := gopacket.NewSerializeBuffer()
	gopacket.SerializeLayers(buf, gopacket.SerializeOptions{
		ComputeChecksums: true,
		FixLengths:       true,
	},
		ls...)

	/*
		_, err := p.p.Write(input.Data())
		if err != nil {
			log.Println("Error writing to Port:", err)
			return
		}

		log.Println("Written!")
	*/
	pkt := gopacket.NewPacket(buf.Bytes(), layers.LayerTypeEthernet, gopacket.DecodeOptions{})
	log.Println("Writing packet:", pkt)
	p.p.PacketSink(gopacket.SerializeOptions{}).NextPacket(pkt)

	consumed = true
	return
}

type portPacketDataSource struct {
	p *Port
}

// Implements gopacket.PacketDataSource
func (src *portPacketDataSource) ReadPacketData() (data []byte, ci gopacket.CaptureInfo, err error) {
	data, n, err := src.p.Read()

	ci.Length = n
	ci.CaptureLength = n
	data = data[:n]

	return
}

func (p *Port) PacketSource(dec gopacket.Decoder) *gopacket.PacketSource {
	src := &portPacketDataSource{
		p: p,
	}
	return gopacket.NewPacketSource(src, dec)
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
