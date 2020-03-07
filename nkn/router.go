package nkn

import (
	"fmt"
	"net"
	"sync"
)


type NknRouter struct {
	lock		sync.Locker
	ipToNknMap	map[string]string
	nknToIpMap	map[string]string
}

func NewNknRouter() *NknRouter {
	r := &NknRouter{
		lock: &sync.Mutex{},
	}
	r.ipToNknMap = make(map[string]string)
	r.nknToIpMap = make(map[string]string)

	return r
}

func (r *NknRouter) AddRoute(ipAddr string, nknAddr string) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	if net.ParseIP(ipAddr) == nil {
		return fmt.Errorf("Cannot add route for invalid IP address: %v", ipAddr)
	}

	r.ipToNknMap[ipAddr] = nknAddr
	r.nknToIpMap[nknAddr] = ipAddr

	return nil
}

func (r *NknRouter) RemoveRoute(ipAddr string, nknAddr string) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	if net.ParseIP(ipAddr) == nil {
		return fmt.Errorf("Cannot add route for invalid IP address: %v", ipAddr)
	}

	actual, ok := r.ipToNknMap[ipAddr]
	if !ok || actual != nknAddr {
		return nil
	}

	delete(r.ipToNknMap, ipAddr)
	delete(r.nknToIpMap, nknAddr)

	return nil
}

func (r *NknRouter) RouteTo(ipAddr string) (nknAddrs []string, err error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	if net.ParseIP(ipAddr) == nil {
		return []string{}, fmt.Errorf("Cannot get route for invalid IP address: %v", ipAddr)
	}

	dest, ok := r.ipToNknMap[ipAddr]
	if !ok {
		return []string{}, fmt.Errorf("No route known for %v", ipAddr)
	}

	return []string{dest}, nil
}