package main

import (
	"errors"
	"fmt"

	"github.com/ainghazal/torii/vpn"
)

func renderConfigForProvider(provider vpn.Provider, selector providerSelectorFn) (*config, error) {
	endpoints := selector(provider)
	if len(endpoints) == 0 {
		return nil, errors.New(errNoConfig)
	}
	auth := provider.Auth()

	netTests := []netTest{}

	for _, endpoint := range endpoints {
		test := netTest{
			TestName: endpoint.Proto, // one of: openvpn, wg
			Inputs: []string{
				fmt.Sprintf(
					"vpn://%s.%s/?addr=%s:%s&transport=%s",
					provider.Name(),
					endpoint.Proto,
					endpoint.IP,
					endpoint.Port,
					endpoint.Transport,
				)},
			Options: vpn.Options{
				Cipher:   "AES-256-GCM",
				Auth:     "SHA512",
				SafeCa:   auth.Ca,
				SafeCert: auth.Cert,
				SafeKey:  auth.Key,
			},
		}
		netTests = append(netTests, test)
	}
	return &config{
		Name:        fmt.Sprintf("openvpn-%s", provider.Name()),
		Description: fmt.Sprintf("measure vpn connection to random %s gateways", provider.Name()),
		Author:      authorName,
		NetTests:    netTests,
	}, nil
}
