package vpn

var (
	customName = "unknown"
)

type CustomProvider struct {
	name       string
	CustomName string
	endpoints  []*Endpoint
	auth       AuthDetails
}

func NewCustomProvider(name string) *CustomProvider {
	endpoints := []*Endpoint{}
	return &CustomProvider{
		name:      name,
		endpoints: endpoints,
	}
}

func (c *CustomProvider) AddEndpoint(e *Endpoint) {
	c.endpoints = append(c.endpoints, e)
}

func (c *CustomProvider) AuthFromProvider(p Provider) bool {
	c.auth = p.Auth()
	return true
}

func (c *CustomProvider) Name() string {
	if c.name != "" {
		return c.name
	}
	return customName
}

func (c *CustomProvider) LongName() string {
	if c.CustomName != "" {
		return c.CustomName
	}
	return c.Name()
}

func (c *CustomProvider) Bootstrap() bool {
	return true
}

// Endpoints returns all the available endpoints.
func (c *CustomProvider) Endpoints() []*Endpoint {
	if c.endpoints == nil {
		return []*Endpoint{}
	}
	return c.endpoints
}

// AuthDetails returns valid authentication for this provider.
func (c *CustomProvider) Auth() AuthDetails {
	return c.auth
}

var _ Provider = &CustomProvider{}
