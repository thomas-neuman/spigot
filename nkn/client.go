package nkn

import (
	"encoding/hex"
	"log"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	sdk "github.com/nknorg/nkn-sdk-go"
)


type NknClient struct {
	account	*sdk.Account
	client	*sdk.Client
	router	*NknRouter
}

func NewNknClient(privateSeed string, localAddr string) (*NknClient, error) {
	seed, err := hex.DecodeString(privateSeed)
	if err != nil {
		return nil, err
	}

	acc, err := sdk.NewAccount(seed)
	if err != nil {
		return nil, err
	}

	client, err := sdk.NewClient(acc, localAddr, nil)
	if err != nil {
		return nil, err
	}

	log.Println("Initialized NKN client", client.Address(), ", waiting to connect...")
	<- client.OnConnect.C
	log.Println("NKN client connected.")

	rtr := NewNknRouter()

	c := &NknClient{
		account: acc,
		client: client,
		router: rtr,
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

		/*
		err = c.client.Send(dests, input.Data())
		if err != nil {
			log.Println("Failed to send NKN message!")
			return
		}
		*/

		log.Println("NKN consumed the packet! :-)")
		consumed = true
	}

	return
}