package nkn

import (
	"log"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	sdk "github.com/nknorg/nkn-sdk-go"
)


type NknClient struct {
	client *sdk.Client
}

func NewNknClient() (*NknClient, error) {
	return nil, nil
}

func (c *NknClient) getNknAddress(ip4 net.IPAddr) ([]string, error) {
	return []string{""}, nil
}

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