package nkn

import (
	"fmt"
	"net"
	"sync"

	"github.com/yl2chen/cidranger"

	. "github.com/thomas-neuman/spigot/config"
)


type NknRoute struct {
	destination	net.IPNet
	nexthop		[]string
}

// Implements cidranger.RangerEntry
func (r *NknRoute) Network() net.IPNet {
	return r.destination
}

type NknRouter struct {
	lock		sync.Locker
	routes	cidranger.Ranger
	ipToNknMap	map[string]string
	nknToIpMap	map[string]string
}

func NewNknRouter(config *Configuration) (*NknRouter, error) {
	r := &NknRouter{
		lock: &sync.Mutex{},
	}
	r.ipToNknMap = make(map[string]string)
	r.nknToIpMap = make(map[string]string)
	r.routes = cidranger.NewPCTrieRanger()

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

	_, subnet, err := net.ParseCIDR(ipAddr)

	if err != nil {
		return fmt.Errorf("Cannot add route for invalid IP address %v: %v", ipAddr, err)
	}

	route := &NknRoute{
		destination: *subnet,
		nexthop: []string{nknAddr},
	}
	r.routes.Insert(route)

	r.ipToNknMap[ipAddr] = nknAddr
	r.nknToIpMap[nknAddr] = ipAddr

	return nil
}

func (r *NknRouter) RemoveRoute(ipAddr string, nknAddr string) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	_, subnet, err := net.ParseCIDR(ipAddr)
	if err != nil {
		return fmt.Errorf("Cannot remove route for invalid IP address %v: %v", ipAddr, err)
	}

	_, _ = r.routes.Remove(*subnet)

	actual, ok := r.ipToNknMap[ipAddr]
	if !ok || actual != nknAddr {
		return nil
	}

	delete(r.ipToNknMap, ipAddr)
	delete(r.nknToIpMap, nknAddr)

	return nil
}

func (r *NknRouter) RouteTo(ipAddr string) (nknAddr []string, err error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	ip := net.ParseIP(ipAddr)
	if ip == nil {
		return nil, fmt.Errorf("Cannot remove route for invalid IP address %v", ipAddr)
	}

	routes, err := r.routes.ContainingNetworks(ip)
	if err != nil {
		return nil, fmt.Errorf("Failed to lookup route for %v: %v", ipAddr, err)
	}

	found := len(routes)
	if found > 0 {
		// Routes are looked up through a trie, and accumulate as we go down through
		// the branches. Therefore, the last entry will be the lowest on the trie, and
		// the most specific route.
		route := routes[found - 1].(*NknRoute)
		return route.nexthop, nil
	}

	return nil, fmt.Errorf("No route known for %v", ipAddr)
}