package port

import (
	"log"

	"github.com/songgao/water"
)


type Port struct {
	Name string
	Iface *water.Interface
}


func (p *Port) Read() (data []byte, len int, err error) {
	data = make([]byte, 1500)
	len, err = p.Iface.Read(data)

	return
}

func (p *Port) Write(data []byte) (len int, err error) {
	return p.Iface.Write(data)
}


func NewPort(name string) (*Port, error) {
	p := &Port{
		Name: name,
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