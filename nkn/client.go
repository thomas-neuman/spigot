package nkn

import (
	"encoding/hex"
	"log"
	"net"

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

func (c *NknClient) getNknAddress(ip4 net.IPAddr) ([]string, error) {
	return []string{""}, nil
}

// Implement packets.PacketProcessor
func (c *NknClient) Process(input gopacket.Packet) (output gopacket.Packet, consumed bool) {
	output = input
	consumed = false

	ip4Layer := input.Layer(layers.LayerTypeIPv4)
	if ip4Layer != nil {
		log.Println("IPv4 message!")

		/*
		dests, err := c.getNknAddress()
		if err != nil {
			log.Println("Could not get destination(s) for NKN message!")
			return
		}

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