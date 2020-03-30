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
	account	*sdk.Account
	client	*sdk.Client
	router	*NknRouter
	msgConf	*sdk.MessageConfig
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

	log.Println("Initialized NKN client", client.Address(), ", waiting to connect...")
	<- client.OnConnect.C
	log.Println("NKN client connected.")

	rtr, err := NewNknRouter(config)
	if err != nil {
		return nil, err
	}

	conf := &sdk.MessageConfig{
		NoReply:	true,
	}

	c := &NknClient{
		account:	acc,
		client:		client,
		router: 	rtr,
		msgConf:	conf,
	}

	return c, nil
}

// Implement packets.PacketProcessor
func (c *NknClient) Process(input gopacket.Packet) (output gopacket.Packet, consumed bool) {
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

		_, err = c.client.Send(dests, input.Data(), c.msgConf)
		if err != nil {
			log.Println("Failed to send NKN message!")
			return
		}

		consumed = true
	}

	return
}