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
	. "github.com/thomas-neuman/spigot/packets"
	. "github.com/thomas-neuman/spigot/port"
)

func processingLoop(ctxt context.Context, src *gopacket.PacketSource, procs ...PacketProcessor) {
	var p gopacket.Packet

	for {
		select {
		case p = <-src.Packets():
			// log.Println("Got frame:", p)
			go func(pkt gopacket.Packet) {
				consumed := false

				for _, proc := range procs {
					pkt, consumed = proc.Process(pkt)
					if consumed {
						return
					}
					// log.Println("    --> Intermediate packet:", frame)
				}

				if !consumed {
					log.Println("Frame was never consumed!")
				}
			}(p)
		case <-ctxt.Done():
			return
		}
	}
}

func main() {
	ctxt := context.Background()
	ctxt, cancel := context.WithCancel(ctxt)
	defer cancel()

	config, err := config.GetConfiguration()
	if err != nil {
		log.Fatal(err)
	}

	br0, err := NewPort(config.IfaceName)
	br0.SetUp(config.IPAddress)
	if err != nil {
		log.Fatal(err)
	}

	// frameSrc := PacketSourceFromPort(br0, layers.LayerTypeEthernet)
	// portSrc := br0.PacketSource(layers.LayerTypeEthernet)

	arpResp := NewArpResponder(br0)

	nknClient, err := nkn.NewNknClient(config, br0, ctxt)
	if err != nil {
		log.Fatal(err)
	}

	go processingLoop(ctxt,
		br0.PacketSource(layers.LayerTypeEthernet),
		arpResp.Egress(),
		nknClient.Egress())
	go processingLoop(ctxt,
		nknClient.PacketSource(layers.LayerTypeIPv4),
		br0.Ingress())

	c := make(chan os.Signal, 1)
	signal.Notify(c,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	// Block until a signal is received.
	s := <-c
	log.Println("Caught signal", s)
}
