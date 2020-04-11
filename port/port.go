package port

import (
	"fmt"
	"log"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/songgao/water"

	"github.com/thomas-neuman/spigot/config"
)

type Port struct {
	conf   *config.Configuration
	iface  *water.Interface
	inbuf  []byte
	outbuf gopacket.SerializeBuffer
	opts   gopacket.SerializeOptions
}

func (p *Port) FirstLayerType() gopacket.LayerType {
	return layers.LayerTypeEthernet
}

func (p *Port) DoInit() error {
	p.inbuf = make([]byte, 1500)

	p.outbuf = gopacket.NewSerializeBuffer()
	p.opts = gopacket.SerializeOptions{
		ComputeChecksums: true,
		FixLengths:       true,
	}
	return nil
}

func (p *Port) DoRead() (data []gopacket.SerializableLayer, err error) {
	_, err = p.iface.Read(p.inbuf)
	if err != nil {
		return
	}

	pkt := gopacket.NewPacket(p.inbuf, p.FirstLayerType(), gopacket.DecodeOptions{})
	for _, l := range pkt.Layers() {
		s, ok := l.(gopacket.SerializableLayer)
		if !ok {
			err = fmt.Errorf("Unusable layer! %v", l)
			return
		}
		data = append(data, s)
	}
	return
}

func (p *Port) DoWrite(data []gopacket.SerializableLayer) (err error) {
	eth := data[0].(*layers.Ethernet)
	eth.SrcMAC = p.HardwareAddr()
	eth.DstMAC = p.HardwareAddr()

	gopacket.SerializeLayers(p.outbuf, p.opts, data...)

	_, err = p.iface.Write(p.outbuf.Bytes())
	if err != nil {
		log.Println("Error writing packet to interface:", err)
	}

	return err
}

func NewPort(conf *config.Configuration) (*Port, error) {
	p := &Port{
		conf: conf,
	}

	log.Println("Creating TAP...")
	ifaceConfig := water.Config{
		DeviceType: water.TAP,
	}
	ifaceConfig.Name = conf.IfaceName

	iface, err := water.New(ifaceConfig)
	if err != nil {
		return nil, err
	}
	p.iface = iface

	return p, nil
}
