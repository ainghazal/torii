package main

import (
	"errors"
	"fmt"

	"github.com/ainghazal/torii/vpn"
)

var providerOptionsMap = map[string]vpn.Options{
	"riseup": vpn.Options{
		Cipher: "AES-256-GCM",
		Auth:   "SHA512",
	},
	"tunnelbear": vpn.Options{
		Cipher:         "AES-256-GCM",
		Auth:           "SHA256",
		Compress:       "comp-lzo-no",
		SafeLocalCreds: true,
	},
}

var providerUsesCertAuth = map[string]bool{
	"riseup":     true,
	"tunnelbear": false,
}

func optionsForProvider(provider string, auth vpn.AuthDetails) vpn.Options {
	opt := providerOptionsMap[provider]
	opt.SafeCa = auth.Ca
	if needsCertAuth, ok := providerUsesCertAuth[provider]; ok && needsCertAuth {
		opt.SafeCert = auth.Cert
		opt.SafeKey = auth.Key
	}
	return opt
}

func renderConfigForProvider(provider vpn.Provider, selector endpointSelectorFn) (*config, error) {
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
					endpoint.Proto,
					provider.Name(),
					endpoint.IP,
					endpoint.Port,
					endpoint.Transport,
				)},
			Options: optionsForProvider(provider.Name(), auth),
		}
		netTests = append(netTests, test)
	}
	return &config{
		Name:        fmt.Sprintf("openvpn-%s", provider.LongName()),
		Description: fmt.Sprintf("measure vpn connection to random %s gateways", provider.LongName()),
		Author:      authorName,
		NetTests:    netTests,
	}, nil
}
