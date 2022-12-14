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
	LongName() string
	Bootstrap() bool
	Endpoints() []*Endpoint
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

// IsIsKnownProvider returns true if the passed provider name is in our list of
// known providers.
func IsKnownProvider(name string) bool {
	for _, provider := range Providers {
		if name == provider.Name() {
			return true
		}
	}
	return false
}
