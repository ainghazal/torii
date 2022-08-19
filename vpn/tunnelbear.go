package vpn

import (
	"log"
)

var (
	tunnelbearName = "tunnelbear"
)

type TunnelbearProvider struct {
	endpoints []*Endpoint
	auth      AuthDetails
}

func (r *TunnelbearProvider) Name() string {
	return tunnelbearName
}

// Bootstrap implements boostrap method. It will fetch endpoitns from tunnelbear
// api, and get a fresh certificate.
func (r *TunnelbearProvider) Bootstrap() bool {
	log.Println("🌱 Bootstrapping Tunnelbear")
	return true
}

// Endpoints returns all the available endpoints.
func (r *TunnelbearProvider) Endpoints() []*Endpoint {
	if r.endpoints == nil {
		return []*Endpoint{}
	}
	return r.endpoints
}

// Endpoitns returns Endpoints filtered by country code.
func (r *TunnelbearProvider) EndpointByCountry(cc string) []*Endpoint {
	return nil
}

// AuthDetails returns valid authentication for this provider.
func (r *TunnelbearProvider) Auth() AuthDetails {
	return r.auth
}

var _ Provider = &TunnelbearProvider{}

const (
	tunnelbearApiURL = ""
)
