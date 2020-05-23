package port

import (
	"fmt"
	"log"
	"net"

	"github.com/vishvananda/netlink"
)

func (p *Port) SetDown() error {
	return nil
}

func (p *Port) SetUp() error {
	cidr, err := netlink.ParseAddr(p.conf.IPAddress)
	if err != nil {
		return err
	}

	li, err := netlink.LinkByName(p.conf.IfaceName)
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

	err = p.AdjustMTU(84)
	if err != nil {
		return err
	}

	return nil
}

func (p *Port) AdjustMTU(dec uint) error {
	li, err := netlink.LinkByName(p.conf.IfaceName)
	if err != nil {
		return err
	}

	log.Println("Adjusting MTU by", dec, "bytes...")
	mtu := li.Attrs().MTU
	mtu = mtu - int(dec)
	if mtu <= 0 {
		return fmt.Errorf("Decrementing MTU resulted in zero frame length!")
	}

	err = netlink.LinkSetMTU(li, mtu)
	if err != nil {
		return err
	}
	log.Println("MTU set to", mtu, ".")

	return nil
}

func (p *Port) HardwareAddr() net.HardwareAddr {
	l, _ := netlink.LinkByName(p.conf.IfaceName)
	return l.Attrs().HardwareAddr
}
