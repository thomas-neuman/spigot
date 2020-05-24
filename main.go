package main

import (
	"context"
	"log"
	_ "net"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/gopacket"

	"github.com/thomas-neuman/spigot/api"
	"github.com/thomas-neuman/spigot/arp"
	"github.com/thomas-neuman/spigot/config"
	"github.com/thomas-neuman/spigot/handler"
	"github.com/thomas-neuman/spigot/nkn"
	"github.com/thomas-neuman/spigot/port"
)

type SpigotDaemon struct {
	ctxt      context.Context
	port      handler.LayerHandler
	arpResp   handler.LayerHandler
	nknClient handler.LayerHandler
}

func NewSpigotDaemon(ctxt context.Context, conf *config.Configuration) *SpigotDaemon {
	br0, err := port.NewPort(conf)
	br0.SetUp()
	if err != nil {
		log.Fatal(err)
	}

	arpResp := arp.NewArpResponder()

	nknClient, err := nkn.NewNknClient(conf)
	if err != nil {
		log.Fatal(err)
	}

	daemon := &SpigotDaemon{
		ctxt:      ctxt,
		port:      handler.NewLayerHandler(br0, ctxt),
		arpResp:   handler.NewLayerHandler(arpResp, ctxt),
		nknClient: handler.NewLayerHandler(nknClient, ctxt),
	}
	return daemon
}

func (d *SpigotDaemon) Start() {
	d.arpResp.Start()
	d.nknClient.Start()
	d.port.Start()

	go d.egressLoop()
	go d.ingressLoop()
}

func (d *SpigotDaemon) ingressLoop() {
	var ls []gopacket.SerializableLayer
	var err error

	for {
		select {
		case <-d.ctxt.Done():
			return
		default:
			ls, err = d.nknClient.Read()
			if err != nil {
				log.Println("Error reading packet data")
				continue
			}

			go d.port.Write(ls)
		}
	}
}

func (d *SpigotDaemon) egressLoop() {
	var ls []gopacket.SerializableLayer
	var err error

	for {
		select {
		case <-d.ctxt.Done():
			return
		default:
			ls, err = d.port.Read()
			if err != nil {
				log.Println("Error reading packet data")
				continue
			}

			go func() {
				for i, l := range ls {
					switch l.LayerType() {
					case d.port.FirstLayerType():
						continue
					case d.arpResp.FirstLayerType():
						err = d.arpResp.Write(ls[i:])
						if err != nil {
							log.Println("Error writing ARP packet to ArpResponder:", err)
							return
						}

						reply, err := d.arpResp.Read()
						if err != nil {
							log.Println("Error reading ARP packet from ArpResponder:", err)
							return
						}

						err = d.port.Write(reply)
						if err != nil {
							log.Println("Error writing ARP back to port:", err)
							return
						}
					case d.nknClient.FirstLayerType():
						err = d.nknClient.Write(ls[i:])
						if err != nil {
							log.Println("Error writing packet to NknClient:", err)
							return
						}
					}
				}
			}()
		}
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.LUTC | log.Lshortfile)
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

	server := api.NewApiServer(ctxt, conf)
	server.Start()

	// Block until a signal is received.
	s := <-c
	log.Println("Caught signal", s)
}
