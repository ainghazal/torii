package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/netip"

	"github.com/ainghazal/torii/vpn"
)

type endpointSelectorFn func(vpn.Provider) []*vpn.Endpoint
type providerFilterFn func(*vpn.Endpoint) bool

func nullFilter(*vpn.Endpoint) bool {
	return true
}

func healthyFilter(provider string) providerFilterFn {
	hs, ok := healthServiceMap[provider]
	if !ok {
		return nullFilter
	}
	return func(endp *vpn.Endpoint) bool {
		addrStr := fmt.Sprintf("%s:%s", endp.IP, endp.Port)
		addr := net.TCPAddrFromAddrPort(netip.MustParseAddrPort(addrStr))
		healthy, err := hs.Healthy(addr, endp.Transport)
		if err != nil {
			log.Println("ERROR:", err)
			return true
		}
		return healthy
	}
}

// filterAndRandomizeEndpointPicker accepts a provider, a boolean filter, and
// an integer indicating the maximum number of desired results. It
// will return an array of pointers to vpn.Endpoint structs, chosen pseudo-randomly after
// applying the passed filter to the list of all endpoints for that provider.
func filterAndRandomizeEndpointsPicker(p vpn.Provider, filter providerFilterFn, max int) (res []*vpn.Endpoint) {
	all := p.Endpoints()
	if len(all) == 0 {
		return nil
	}
	sel := []*vpn.Endpoint{}
	healthy := healthyFilter(p.Name())
	for _, endp := range all {
		if filter(endp) && healthy(endp) {
			sel = append(sel, endp)
		}
	}
	if len(sel) == 0 {
		return res
	}
	for i := 0; i < max; i++ {
		pick := rand.Intn(len(sel))
		log.Printf("🎲 Picked endpoint %d/%d\n", pick+1, len(sel))
		res = append(res, sel[pick])
	}
	return res
}

// randomEndpointPicker returns a provider selector that picks one random
// endpoint.
func randomEndpointPicker() endpointSelectorFn {
	all := func(e *vpn.Endpoint) bool {
		return true
	}
	// curry filterAndRandomizeEndpointPicker
	return func(p vpn.Provider) []*vpn.Endpoint {
		return filterAndRandomizeEndpointsPicker(p, all, 1)
	}
}

// byCountryEndpointPicker returns a provider selector that picks a number max
// of endpoints after filtering by country code.
func byCountryEndpointPicker(cc string, max int) endpointSelectorFn {
	filterByCC := func(e *vpn.Endpoint) bool {
		if e.CountryCode == cc {
			return true
		}
		return false
	}
	// curry filterAndRandomizeEndpointPicker
	return func(p vpn.Provider) []*vpn.Endpoint {
		return filterAndRandomizeEndpointsPicker(p, filterByCC, max)
	}
}
