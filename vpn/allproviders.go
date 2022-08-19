package vpn

// Endpoint is a single instance of any remote endpoint for a VPN Connection.
type Endpoint struct {
	Label       string
	IP          string
	Port        string
	Proto       string
	Transport   string
	Obfuscation string
	CountryCode string
}

// Provider is the entity that runs endpoints.
type Provider interface {
	Name() string
	Bootstrap() bool
	Endpoints() []*Endpoint
	EndpointByCountry(string) []*Endpoint
	Auth() AuthDetails
}

// AuthDetails are generic credentials needed to authenticate with an endpoint.
// At the moment only certificates are supported.
type AuthDetails struct {
	Ca   string
	Cert string
	Key  string
}

// Providers is a map that allows to select providers by their name.
var Providers = map[string]Provider{
	"riseup":     &RiseupProvider{},
	"tunnelbear": &TunnelbearProvider{},
}

// InitAllProviders calls the Bootstrap method on all the registered providers.
func InitAllProviders() error {
	for _, provider := range Providers {
		provider.Bootstrap()
	}
	return nil
}

func IsKnownProvider(name string) bool {
	for _, provider := range Providers {
		if name == provider.Name() {
			return true
		}
	}
	return false
}
