package port

import (
	"log"
	"net"

	"github.com/vishvananda/netlink"
)


func (p *Port) SetDown() error {
	return nil
}

func (p *Port) SetUp(addr string) error {
	cidr, err := netlink.ParseAddr(addr)
	if err != nil {
		return err
	}
	// p.Cidr = *cidr

	li, err := netlink.LinkByName(p.Name)
	if err != nil {
		return err
	}
	err = netlink.LinkSetUp(li)
	if err != nil {
		return err
	}

	log.Println("Setting TAP address...")
	err = netlink.AddrAdd(li, cidr)
	if err != nil {
		return err
	}
	log.Println("TAP address set.")

	return nil
}

func (p *Port) HardwareAddr() net.HardwareAddr {
	l, _ := netlink.LinkByName(p.Name)
	return l.Attrs().HardwareAddr
}