package nkn

import (
	"encoding/hex"
	"io/ioutil"
	"log"
	"strings"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	sdk "github.com/nknorg/nkn-sdk-go"

	. "github.com/thomas-neuman/spigot/config"
)

type NknClient struct {
	account *sdk.Account
	client  *sdk.Client
	router  *NknRouter
	msgConf *sdk.MessageConfig
	outbuf  gopacket.SerializeBuffer
	opts    gopacket.SerializeOptions
}

func NewNknClient(config *Configuration) (*NknClient, error) {
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
	}

	return c, nil
}

func (c *NknClient) FirstLayerType() gopacket.LayerType {
	return layers.LayerTypeIPv4
}

func (c *NknClient) DoInit() error {
	c.outbuf = gopacket.NewSerializeBuffer()
	c.opts = gopacket.SerializeOptions{
		ComputeChecksums: true,
		FixLengths:       true,
	}

	log.Println("Waiting for client to connect...")
	<-c.client.OnConnect.C
	log.Println("Connected.")

	return nil
}

func (c *NknClient) DoRead() (data []gopacket.SerializableLayer, err error) {
	msg := <-c.client.OnMessage.C

	pkt := gopacket.NewPacket(msg.Data, c.FirstLayerType(), gopacket.DecodeOptions{})

	data = []gopacket.SerializableLayer{
		&layers.Ethernet{
			EthernetType: layers.EthernetTypeIPv4,
		},
	}
	for _, l := range pkt.Layers() {
		data = append(data, l.(gopacket.SerializableLayer))
	}

	return
}

func (c *NknClient) DoWrite(data []gopacket.SerializableLayer) (err error) {
	ip4 := data[0].(*layers.IPv4)

	dests, err := c.router.RouteTo(ip4.DstIP.String())
	if err != nil {
		log.Println("Could not get destination(s) for NKN message!", err)
		return
	}

	gopacket.SerializeLayers(c.outbuf, c.opts, data...)
	_, err = c.client.Send(dests, c.outbuf.Bytes(), c.msgConf)
	return
}
