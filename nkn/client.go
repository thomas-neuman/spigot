package nkn

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	sdk "github.com/nknorg/nkn-sdk-go"
	"github.com/nknorg/nkn/util/address"

	. "github.com/thomas-neuman/spigot/config"
)

type NknClient struct {
	account *sdk.Account
	client  *sdk.Client
	router  *NknRouter
	msgConf *sdk.MessageConfig

	authorizedKeys map[string]bool

	outbuf gopacket.SerializeBuffer
	opts   gopacket.SerializeOptions
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

	keys := make(map[string]bool)
	keys[string(acc.PubKey())] = true

	for _, k := range config.AuthorizedKeys {
		keys[k] = true
	}
	for _, r := range config.StaticRoutes {
		_, pubkey, _, err := address.ParseClientAddress(r.Nexthop)
		if err != nil {
			log.Println("Error parsing nexthop address:", err)
		}
		keys[string(pubkey)] = true
	}

	c := &NknClient{
		account:        acc,
		client:         client,
		router:         rtr,
		msgConf:        conf,
		authorizedKeys: keys,
	}

	return c, nil
}

func (c *NknClient) FirstLayerType() gopacket.LayerType {
	return layers.LayerTypeIPv4
}

func (c *NknClient) DoInit() error {
	c.outbuf = gopacket.NewSerializeBuffer()
	c.opts = gopacket.SerializeOptions{}

	log.Println("Waiting for client to connect...")
	<-c.client.OnConnect.C
	log.Println("Connected.")

	return nil
}

func (c *NknClient) isMessageAuthorized(msg *sdk.Message) bool {
	src := msg.Src
	_, pubkey, _, err := address.ParseClientAddress(src)
	if err != nil {
		log.Println("Error parsing source address:", err)
		return false
	}

	authorized, ok := c.authorizedKeys[string(pubkey)]
	return (ok && authorized)
}

func (c *NknClient) DoRead() (data []gopacket.SerializableLayer, err error) {
	msg := <-c.client.OnMessage.C

	if !(c.isMessageAuthorized(msg)) {
		err = errors.New("Message not authorized!")
		return
	}

	pkt := gopacket.NewPacket(msg.Data, c.FirstLayerType(), gopacket.DecodeOptions{})

	data = []gopacket.SerializableLayer{
		&layers.Ethernet{
			EthernetType: layers.EthernetTypeIPv4,
		},
	}
	for _, l := range pkt.Layers() {
		s, ok := l.(gopacket.SerializableLayer)
		if !ok {
			log.Printf("Cannot read layer as SerializableLayer: %v", l)
			err = fmt.Errorf("Unusable layer! %v", l)
			return
		}
		data = append(data, s)
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
