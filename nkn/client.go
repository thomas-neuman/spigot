package nkn

import (
	"context"
	"encoding/hex"
	"io/ioutil"
	"log"
	"strings"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	sdk "github.com/nknorg/nkn-sdk-go"

	. "github.com/thomas-neuman/spigot/config"
	"github.com/thomas-neuman/spigot/packets"
	. "github.com/thomas-neuman/spigot/port"
)

type NknClient struct {
	account *sdk.Account
	client  *sdk.Client
	router  *NknRouter
	msgConf *sdk.MessageConfig
	context context.Context
	snk     packets.PacketSink
}

func NewNknClient(config *Configuration, port *Port, ctxt context.Context) (*NknClient, error) {
	seed, err := ioutil.ReadFile(config.PrivateSeedFile)
	if err != nil {
		log.Fatal(err)
	}

	hexSeed, err := hex.DecodeString(strings.TrimSpace(string(seed)))
	if err != nil {
		return nil, err
	}

	acc, err := sdk.NewAccount(hexSeed)
	if err != nil {
		return nil, err
	}

	client, err := sdk.NewClient(acc, config.IPAddress, nil)
	if err != nil {
		return nil, err
	}

	log.Println("Initialized NKN client", client.Address(), ", waiting to connect...")
	<-client.OnConnect.C
	log.Println("NKN client connected.")

	rtr, err := NewNknRouter(config)
	if err != nil {
		return nil, err
	}

	conf := &sdk.MessageConfig{
		NoReply: true,
	}

	c := &NknClient{
		account: acc,
		client:  client,
		router:  rtr,
		msgConf: conf,
		context: ctxt,
		snk:     port.PacketSink(gopacket.SerializeOptions{}),
	}

	return c, nil
}

type nknClientPacketDataSource struct {
	c *NknClient
}

// Implements gopacket.PacketDataSource
func (src *nknClientPacketDataSource) ReadPacketData() (data []byte, ci gopacket.CaptureInfo, err error) {
	om := src.c.client.OnMessage.C
	done := src.c.context.Done()

	var msg *sdk.Message

	select {
	case msg = <-om:
		log.Printf("Got message: %v", msg)

		data = msg.Data

		n := len(data)
		ci.Length = n
		ci.CaptureLength = n

		pkt := gopacket.NewPacket(data, layers.LayerTypeEthernet, gopacket.DecodeOptions{})
		err = src.c.snk.NextPacket(pkt)
		return
	case <-done:
		return
	}
}

func (c *NknClient) PacketSource(dec gopacket.Decoder) *gopacket.PacketSource {
	src := &nknClientPacketDataSource{
		c: c,
	}
	return gopacket.NewPacketSource(src, dec)
}

func (c *NknClient) Egress() *nknClientEgress {
	return &nknClientEgress{
		c: c,
	}
}

type nknClientEgress struct {
	c *NknClient
}

// Implement packets.PacketProcessor
func (ce *nknClientEgress) Process(input gopacket.Packet) (output gopacket.Packet, consumed bool) {
	c := ce.c
	output = input
	consumed = false

	ip4Layer := input.Layer(layers.LayerTypeIPv4)
	if ip4Layer != nil {
		log.Println("IPv4 message!")

		ip4 := ip4Layer.(*layers.IPv4)

		dests, err := c.router.RouteTo(ip4.DstIP.String())
		if err != nil {
			log.Println("Could not get destination(s) for NKN message!", err)
			return
		}

		log.Println("Got destination(s):", dests)

		_, err = c.client.Send(dests, input.LinkLayer().LayerPayload(), c.msgConf)
		if err != nil {
			log.Println("Failed to send NKN message!")
			return
		}

		consumed = true
	}

	return
}

func (c *NknClient) Send(dests *sdk.StringArray, payload []byte) error {
	_, err := c.client.Send(dests, payload, c.msgConf)
	return err
}
