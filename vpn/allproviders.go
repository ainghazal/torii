package vpn

// Endpoint is a single instance of any remote endpoint for a VPN Connection.
type Endpoint struct {
	IP          string
	Port        string
	Proto       string
	Transport   string
	Obfuscation string
}

// Provider is the entity that runs endpoints.
type Provider interface {
	Name() string
	Bootstrap() bool
	Endpoints() []*Endpoint
	EndpointByCountry(string) []*Endpoint
	Auth() AuthDetails
}

type AuthDetails struct {
	Ca   string
	Cert string
	Key  string
}

// Providers is a map that allows to select providers by their name.
var Providers = map[string]Provider{
	"riseup": &RiseupProvider{},
}
