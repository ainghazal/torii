package share

type experiment struct {
	ID           int    `json:"ID"`
	Name         string `json:"name"`
	Provider     string `json:"provider"`
	CountryCode  string `json:"cc"`
	Comment      string `json:"comment"`
	Max          string `json:"max"`
	EndpointHost string
	UUID         string
}

type result struct {
	OK   bool   `json:"ok"`
	Data string `json:"data"`
}

type resultExp struct {
	OK   bool          `json:"ok"`
	Data []*experiment `json:"data"`
}
