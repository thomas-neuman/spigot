package nkn

import (
	"fmt"
	"net"
)


type NknRouter struct {
	ipToNknMap map[string]string
	nknToIpMap map[string]string
}

func NewNknRouter() *NknRouter {
	r := &NknRouter{}
	r.ipToNknMap = make(map[string]string)
	r.nknToIpMap = make(map[string]string)

	return r
}

func (r *NknRouter) AddRoute(ipAddr string, nknAddr string) error {
	if net.ParseIP(ipAddr) == nil {
		return fmt.Errorf("Cannot add route for invalid IP address: %v", ipAddr)
	}

	r.ipToNknMap[ipAddr] = nknAddr
	r.nknToIpMap[nknAddr] = ipAddr

	return nil
}