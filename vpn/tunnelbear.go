package vpn

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	tunnelbearName         = "tunnelbear"
	tunnelbearApiConfigURL = "https://tunnelbear.s3.amazonaws.com/support/linux/openvpn.zip"

	configFileName = "openvpn.zip"
)

type countryCodeDomainMap = map[string]string

type TunnelbearProvider struct {
	endpoints []*Endpoint
	auth      AuthDetails
	domainMap countryCodeDomainMap
}

func (t *TunnelbearProvider) Name() string {
	return tunnelbearName
}

// Bootstrap implements boostrap method. It will fetch endpoitns from tunnelbear
// api, and get a fresh certificate.
func (t *TunnelbearProvider) Bootstrap() bool {
	log.Println("🌱 Bootstrapping Tunnelbear")
	if !hasConfigZipFile() {
		downloadAndExtractConfigFile(tunnelbearApiConfigURL)
	}
	t.domainMap = extractCountryDomainsFromConfigFolder(openVPNConfigPath())
	log.Printf("-- Got %d endpoint domains\n", len(t.domainMap))

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

func tunnelBearPath() string {
	return filepath.Join(".", "data", "tunnelbear")
}

func hasConfigZipFile() bool {
	fn := filepath.Join(tunnelBearPath(), configFileName)
	if _, err := os.Stat(fn); errors.Is(err, os.ErrNotExist) {
		return false
	} else {
		return true
	}
}

func openVPNConfigPath() string {
	return filepath.Join(tunnelBearPath(), "config", "openvpn")
}

func downloadAndExtractConfigFile(uri string) {
	os.MkdirAll(tunnelBearPath(), os.ModePerm)
	// Create blank file
	fn := filepath.Join(tunnelBearPath(), configFileName)
	file, err := os.Create(fn)
	if err != nil {
		log.Fatal(err)
	}
	client := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}
	// Put content on file
	resp, err := client.Get(uri)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	size, err := io.Copy(file, resp.Body)
	defer file.Close()

	log.Printf("Downloaded config file %s with size %d\n", uri, size)

	unzip(file.Name(), filepath.Join(tunnelBearPath(), "config"))
}

func extractCountryDomainsFromConfigFolder(path string) countryCodeDomainMap {
	dm := make(map[string]string)
	lines := findInDir("remote", []string{path})
	for _, line := range lines {
		words := strings.Split(line, " ")
		if words[0] != "remote" {
			continue
		}
		domain := words[1]
		port := words[2]

		cc := getCountryCodeFromSubdomain(domain)

		remote := fmt.Sprintf("%s:%s", domain, port)
		dm[cc] = remote
	}
	return dm
}

func getCountryCodeFromSubdomain(d string) string {
	p := strings.Split(d, ".")
	if len(p) == 0 {
		return ""
	}
	return p[0]
}
