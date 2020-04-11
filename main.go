package main

import (
	"context"
	"log"
	_ "net"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	_ "github.com/songgao/packets/ethernet"
	_ "github.com/vishvananda/netlink"

	. "github.com/thomas-neuman/spigot/arp"
	"github.com/thomas-neuman/spigot/config"
	"github.com/thomas-neuman/spigot/nkn"
	_ "github.com/thomas-neuman/spigot/packets"
	. "github.com/thomas-neuman/spigot/port"
)

type SpigotDaemon struct {
	ctxt      context.Context
	port      *Port
	arpResp   *ArpResponder
	nknClient *nkn.NknClient

	ingressInput chan []byte
	egressInput  chan []byte
}

func NewSpigotDaemon(ctxt context.Context, conf *config.Configuration) *SpigotDaemon {
	br0, err := NewPort(conf.IfaceName)
	br0.SetUp(conf.IPAddress)
	if err != nil {
		log.Fatal(err)
	}

	arpResp := NewArpResponder(br0)

	nknClient, err := nkn.NewNknClient(conf, br0, ctxt)
	if err != nil {
		log.Fatal(err)
	}

	daemon := &SpigotDaemon{
		ctxt:      ctxt,
		port:      br0,
		arpResp:   arpResp,
		nknClient: nknClient,
	}
	return daemon
}

func (d *SpigotDaemon) Start() {
	go d.egressLoop()
}

func (d *SpigotDaemon) egressLoop() {
	var eth layers.Ethernet
	var arp layers.ARP
	var ip4 layers.IPv4
	parser := gopacket.NewDecodingLayerParser(layers.LayerTypeEthernet, &eth, &arp, &ip4)
	parser.IgnoreUnsupported = true
	dec := []gopacket.LayerType{}

	var b []byte
	var err error
	for {
		select {
		case <-d.ctxt.Done():
			return
		default:
			b, _, err = d.port.Read()
			if err != nil {
				log.Println("Error reading packet data")
				continue
			}

			err = parser.DecodeLayers(b, &dec)
			if err != nil {
				log.Println("Error decoding egress packet:", err)
				continue
			}

			for _, lt := range dec {
				switch lt {
				case layers.LayerTypeEthernet:
					continue
				case layers.LayerTypeARP:
					err = d.arpResp.Egress().Process(&arp)
				case layers.LayerTypeIPv4:
					d.nknClient.Egress().Process(&ip4)
				}
			}
		}
	}
}

func main() {
	ctxt := context.Background()
	ctxt, cancel := context.WithCancel(ctxt)
	defer cancel()

	conf, err := config.GetConfiguration()
	if err != nil {
		log.Fatal(err)
	}

	daemon := NewSpigotDaemon(ctxt, conf)

	c := make(chan os.Signal, 1)
	signal.Notify(c,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	daemon.Start()

	// Block until a signal is received.
	s := <-c
	log.Println("Caught signal", s)
}
