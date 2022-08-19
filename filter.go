package main

import (
	"log"
	"math/rand"

	"github.com/ainghazal/torii/vpn"
)

type providerSelectorFn func(vpn.Provider) *vpn.Endpoint
type providerFilterFn func(*vpn.Endpoint) bool

// filterAndRandomizeEndpointPicker accepts a provider and a boolean filter. It
// will return a pointer to a vpn.Endpoint struct, chosen pseudo-randomly after
// applying the passed filter to the list of all endpoints for that provider.
func filterAndRandomizeEndpointPicker(p vpn.Provider, filter providerFilterFn) *vpn.Endpoint {
	all := p.Endpoints()
	if len(all) == 0 {
		return nil
	}
	sel := []*vpn.Endpoint{}
	for _, endp := range all {
		if filter(endp) {
			sel = append(sel, endp)
		}
	}
	if len(sel) == 0 {
		return nil
	}
	pick := rand.Intn(len(sel))
	log.Printf("🎲 Picked endpoint %d/%d\n", pick+1, len(sel))
	return sel[pick]
}

func randomEndpointPicker() providerSelectorFn {
	all := func(e *vpn.Endpoint) bool {
		return true
	}
	// curry filterAndRandomizeEndpointPicker
	return func(p vpn.Provider) *vpn.Endpoint {
		return filterAndRandomizeEndpointPicker(p, all)
	}
}

func byCountryEndpointPicker(cc string) providerSelectorFn {
	filterByCC := func(e *vpn.Endpoint) bool {
		if e.CountryCode == cc {
			return true
		}
		return false
	}
	// curry filterAndRandomizeEndpointPicker
	return func(p vpn.Provider) *vpn.Endpoint {
		return filterAndRandomizeEndpointPicker(p, filterByCC)
	}
}
