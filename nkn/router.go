package nkn

import (
	"fmt"
	"net"
	"sync"

	sdk "github.com/nknorg/nkn-sdk-go"
	"github.com/nknorg/nkn/util/address"

	. "github.com/thomas-neuman/spigot/config"
)

type NknRouter struct {
	lock       sync.Locker
	ipToNknMap map[string]string
	nknToIpMap map[string]string
}

func NewNknRouter(config *Configuration) (*NknRouter, error) {
	r := &NknRouter{
		lock: &sync.Mutex{},
	}
	r.ipToNknMap = make(map[string]string)
	r.nknToIpMap = make(map[string]string)

	for _, rr := range config.StaticRoutes {
		err := r.AddRoute(rr.Destination, rr.Nexthop)
		if err != nil {
			return nil, err
		}
	}

	return r, nil
}

func (r *NknRouter) AddRoute(ipAddr string, nknAddr string) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	ip := net.ParseIP(ipAddr)
	if ip == nil {
		return fmt.Errorf("Cannot add route for invalid IP address %v", ipAddr)
	}

	_, _, _, err := address.ParseClientAddress(nknAddr)
	if err != nil {
		return fmt.Errorf("Error parsing nexthop address: %v", err)
	}

	r.ipToNknMap[ipAddr] = nknAddr
	r.nknToIpMap[nknAddr] = ipAddr

	return nil
}

func (r *NknRouter) RemoveRoute(ipAddr string, nknAddr string) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	ip := net.ParseIP(ipAddr)
	if ip == nil {
		return fmt.Errorf("Cannot remove route for invalid IP address %v", ipAddr)
	}

	_, _, _, err := address.ParseClientAddress(nknAddr)
	if err != nil {
		return fmt.Errorf("Error parsing nexthop address: %v", err)
	}

	actual, ok := r.ipToNknMap[ipAddr]
	if !ok || actual != nknAddr {
		return nil
	}

	delete(r.ipToNknMap, ipAddr)
	delete(r.nknToIpMap, nknAddr)

	return nil
}

func (r *NknRouter) RouteTo(ipAddr string) (dests *sdk.StringArray, err error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	ip := net.ParseIP(ipAddr)
	if ip == nil {
		return nil, fmt.Errorf("Cannot remove route for invalid IP address %v", ipAddr)
	}

	addrs, ok := r.ipToNknMap[ipAddr]
	if !ok {
		return nil, fmt.Errorf("No route known for %v", ipAddr)
	}

	return sdk.NewStringArray(addrs), nil
}
