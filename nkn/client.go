package nkn

import (
	"context"
	"encoding/hex"
	"errors"
	"io/ioutil"
	"log"
	"strings"
	"sync"

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

func (c *NknClient) Start() {
	om := c.client.OnMessage.C
	done := c.context.Done()

	var msg *sdk.Message

	for {
		select {
		case msg = <-om:
			log.Printf("Got message: %v", msg)

			// c.snk.NextPacket(msg.Data)
		case <-done:
			return
		}
	}
}

func (c *NknClient) Read() (data []byte, n int, err error) {
	select {
	case msg := <-c.client.OnMessage.C:
		log.Printf("Got message: %v", msg)

		data = msg.Data
		n = len(data)
		err = nil

		return
	case <-c.context.Done():
		err = errors.New("Done!")
		return
	}
}

func (c *NknClient) Egress() *nknClientEgress {
	return &nknClientEgress{
		c: c,
	}
}

type nknClientEgress struct {
	c       *NknClient
	buf     gopacket.SerializeBuffer
	opts    gopacket.SerializeOptions
	bufInit sync.Once
}

// Implement packets.PacketProcessor
func (ce *nknClientEgress) Process(input *layers.IPv4) error {
	ce.bufInit.Do(func() {
		ce.buf = gopacket.NewSerializeBuffer()
		ce.opts = gopacket.SerializeOptions{
			FixLengths:       true,
			ComputeChecksums: true,
		}
	})
	c := ce.c

	dests, err := c.router.RouteTo(input.DstIP.String())
	if err != nil {
		log.Println("Could not get destination(s) for NKN message!", err)
		return errors.New("AAAAAAAAAAAHHHH")
	}

	log.Println("Got destination(s):", dests)

	input.SerializeTo(ce.buf, ce.opts)
	gopacket.SerializeLayers(ce.buf, ce.opts, input)
	b, err := ce.buf.AppendBytes(len(input.LayerPayload()))
	if err != nil {
		log.Println("Allocation issue")
		return errors.New("Yeah")
	}
	copy(b, input.LayerPayload())

	log.Println("Sending", ce.buf.Bytes())
	_, err = c.client.Send(dests, ce.buf.Bytes(), c.msgConf)
	if err != nil {
		log.Println("Failed to send NKN message!")
		return errors.New("AAAAAAAAAAAHHHH")
	}

	return nil
}

func (c *NknClient) Send(dests *sdk.StringArray, payload []byte) error {
	_, err := c.client.Send(dests, payload, c.msgConf)
	return err
}
