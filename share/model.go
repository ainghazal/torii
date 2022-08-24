package share

type Experiment struct {
	ID             int    `json:"ID"`
	Name           string `json:"name"`
	Provider       string `json:"provider"`
	CountryCode    string `json:"cc"`
	Comment        string `json:"comment"`
	Max            string `json:"max"`
	EndpointRemote string `json:"endpoint_remote"`
	UUID           string
}

type result struct {
	OK   bool   `json:"ok"`
	Data string `json:"data"`
}

type resultExp struct {
	OK   bool          `json:"ok"`
	Data []*Experiment `json:"data"`
}
