package share

type experiment struct {
	ID           int
	Name         string `json:"name"`
	Provider     string `json:"provider"`
	CountryCode  string `json:"cc"`
	Comment      string `json:"comment"`
	EndpointHost string
	UUID         string
}

type result struct {
	OK bool `json:"ok"`
}
