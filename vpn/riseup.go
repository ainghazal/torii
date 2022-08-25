package vpn

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

var (
	riseupName = "riseup"
	// port 53 is proving to be problematic
	portsToAvoid = []int{53}
)

type RiseupProvider struct {
	endpoints []*Endpoint
	auth      AuthDetails
}

func (r *RiseupProvider) Name() string {
	return riseupName
}

func (r *RiseupProvider) LongName() string {
	return riseupName
}

// Bootstrap implements boostrap method. It will fetch endpoitns from riseup
// api, and get a fresh certificate.
// TODO this certificate needs to be refreshed every few days.
func (r *RiseupProvider) Bootstrap() bool {
	log.Println("🌱 Bootstrapping Riseup")
	endp, err := fetchEndpointsFromAPI()
	if err != nil {
		log.Println("error parsing endpoints:", err)
		return false
	}
	r.endpoints = endp
	log.Printf("-- Got %d endpoint combinations\n", len(endp))
	auth, err := fetchCertificateFromAPI()
	if err != nil {
		log.Println(err)
		return false
	}
	r.auth = auth
	return true
}

// Endpoints returns all the available endpoints.
func (r *RiseupProvider) Endpoints() []*Endpoint {
	if r.endpoints == nil {
		return []*Endpoint{}
	}
	return r.endpoints
}

// AuthDetails returns valid authentication for this provider.
func (r *RiseupProvider) Auth() AuthDetails {
	return r.auth
}

var _ Provider = &RiseupProvider{}

const (
	apiURL  = "https://api.black.riseup.net/3/config/eip-service.json"
	certURL = "https://api.black.riseup.net/3/cert"
)

func getClient() *http.Client {
	root := x509.NewCertPool()
	root.AppendCertsFromPEM(riseupCA)
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
				RootCAs:            root,
			},
		},
	}
	return client
}

func doGet(uri string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return []byte{}, err
	}
	resp, err := getClient().Do(req)
	if err != nil {
		return []byte{}, err
	}
	if resp.StatusCode != 200 {
		return []byte{}, fmt.Errorf("err code: %s", resp.Status)
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func fetchEndpointsFromAPI() ([]*Endpoint, error) {
	endp := []*Endpoint{}
	eipJson, _ := doGet(apiURL)
	eip := &eipService{}
	err := json.Unmarshal(eipJson, eip)
	if err != nil {
		return endp, err
	}
	for _, gw := range eip.Gateways {
		ipaddr := gw.IPAddress
		label := gw.Host
		cc := eip.Locations[gw.Location].CountryCode
		for _, transport := range gw.Capabilities.Transport {
			for _, proto := range transport.Protocols {
				for _, port := range transport.Ports {

					if shouldAvoidPort(port) {
						continue
					}

					var obfs string
					switch transport.Type {
					case "obfs4":
						obfs = "obfs4"
					default:
						obfs = "none"
					}
					e := &Endpoint{
						Label:       label,
						IP:          ipaddr,
						Port:        port,
						Proto:       "openvpn",
						Transport:   proto, // the semantics are switched here
						CountryCode: strings.ToLower(cc),
						Obfuscation: obfs,
					}
					endp = append(endp, e)
				}
			}
		}
	}
	return endp, nil
}

func fetchCertificateFromAPI() (AuthDetails, error) {
	cert, err := doGet(certURL)
	if err != nil {
		return AuthDetails{}, err
	}
	key, crt, err := splitCombinedPEM(cert)
	if err != nil {
		return AuthDetails{}, err
	}
	auth := AuthDetails{
		Key:  string(toBase64(key)),
		Cert: string(toBase64(crt)),
		Ca:   string(toBase64(riseupVPNCA)),
	}
	return auth, nil
}

func shouldAvoidPort(port string) bool {
	i, err := strconv.Atoi(port)
	if err != nil {
		log.Printf("WARN bad port %v", i)
		return true
	}
	for _, p := range portsToAvoid {
		if p == i {
			return true
		}
	}
	return false
}

//
// Data structures to parse Riseup VPN API
//

type eipService struct {
	Gateways             []gateway
	Locations            map[string]Location
	OpenvpnConfiguration openvpnConfig `json:"openvpn_configuration"`
}

type openvpnConfig map[string]interface{}

type gateway struct {
	Capabilities struct {
		Transport []transport
	}
	Host      string
	IPAddress string `json:"ip_address"`
	Location  string
}

type Location struct {
	CountryCode string `json:"country_code"`
	Hemisphere  string
	Name        string
	Timezone    string
}

type transport struct {
	Type      string
	Protocols []string
	Ports     []string
	Options   map[string]string
}

var riseupCA = []byte(`
-----BEGIN CERTIFICATE-----
MIIFjTCCA3WgAwIBAgIBATANBgkqhkiG9w0BAQ0FADBZMRgwFgYDVQQKDA9SaXNl
dXAgTmV0d29ya3MxGzAZBgNVBAsMEmh0dHBzOi8vcmlzZXVwLm5ldDEgMB4GA1UE
AwwXUmlzZXVwIE5ldHdvcmtzIFJvb3QgQ0EwHhcNMTQwNDI4MDAwMDAwWhcNMjQw
NDI4MDAwMDAwWjBZMRgwFgYDVQQKDA9SaXNldXAgTmV0d29ya3MxGzAZBgNVBAsM
Emh0dHBzOi8vcmlzZXVwLm5ldDEgMB4GA1UEAwwXUmlzZXVwIE5ldHdvcmtzIFJv
b3QgQ0EwggIiMA0GCSqGSIb3DQEBAQUAA4ICDwAwggIKAoICAQC76J4ciMJ8Sg0m
TP7DF2DT9zNe0Csk4myoMFC57rfJeqsAlJCv1XMzBmXrw8wq/9z7XHv6n/0sWU7a
7cF2hLR33ktjwODlx7vorU39/lXLndo492ZBhXQtG1INMShyv+nlmzO6GT7ESfNE
LliFitEzwIegpMqxCIHXFuobGSCWF4N0qLHkq/SYUMoOJ96O3hmPSl1kFDRMtWXY
iw1SEKjUvpyDJpVs3NGxeLCaA7bAWhDY5s5Yb2fA1o8ICAqhowurowJpW7n5ZuLK
5VNTlNy6nZpkjt1QycYvNycffyPOFm/Q/RKDlvnorJIrihPkyniV3YY5cGgP+Qkx
HUOT0uLA6LHtzfiyaOqkXwc4b0ZcQD5Vbf6Prd20Ppt6ei0zazkUPwxld3hgyw58
m/4UIjG3PInWTNf293GngK2Bnz8Qx9e/6TueMSAn/3JBLem56E0WtmbLVjvko+LF
PM5xA+m0BmuSJtrD1MUCXMhqYTtiOvgLBlUm5zkNxALzG+cXB28k6XikXt6MRG7q
hzIPG38zwkooM55yy5i1YfcIi5NjMH6A+t4IJxxwb67MSb6UFOwg5kFokdONZcwj
shczHdG9gLKSBIvrKa03Nd3W2dF9hMbRu//STcQxOailDBQCnXXfAATj9pYzdY4k
ha8VCAREGAKTDAex9oXf1yRuktES4QIDAQABo2AwXjAdBgNVHQ4EFgQUC4tdmLVu
f9hwfK4AGliaet5KkcgwDgYDVR0PAQH/BAQDAgIEMAwGA1UdEwQFMAMBAf8wHwYD
VR0jBBgwFoAUC4tdmLVuf9hwfK4AGliaet5KkcgwDQYJKoZIhvcNAQENBQADggIB
AGzL+GRnYu99zFoy0bXJKOGCF5XUXP/3gIXPRDqQf5g7Cu/jYMID9dB3No4Zmf7v
qHjiSXiS8jx1j/6/Luk6PpFbT7QYm4QLs1f4BlfZOti2KE8r7KRDPIecUsUXW6P/
3GJAVYH/+7OjA39za9AieM7+H5BELGccGrM5wfl7JeEz8in+V2ZWDzHQO4hMkiTQ
4ZckuaL201F68YpiItBNnJ9N5nHr1MRiGyApHmLXY/wvlrOpclh95qn+lG6/2jk7
3AmihLOKYMlPwPakJg4PYczm3icFLgTpjV5sq2md9bRyAg3oPGfAuWHmKj2Ikqch
Td5CHKGxEEWbGUWEMP0s1A/JHWiCbDigc4Cfxhy56CWG4q0tYtnc2GMw8OAUO6Wf
Xu5pYKNkzKSEtT/MrNJt44tTZWbKV/Pi/N2Fx36my7TgTUj7g3xcE9eF4JV2H/sg
tsK3pwE0FEqGnT4qMFbixQmc8bGyuakr23wjMvfO7eZUxBuWYR2SkcP26sozF9PF
tGhbZHQVGZUTVPyvwahMUEhbPGVerOW0IYpxkm0x/eaWdTc4vPpf/rIlgbAjarnJ
UN9SaWRlWKSdP4haujnzCoJbM7dU9bjvlGZNyXEekgeT0W2qFeGGp+yyUWw8tNsp
0BuC1b7uW/bBn/xKm319wXVDvBgZgcktMolak39V7DVO
-----END CERTIFICATE-----
-----BEGIN CERTIFICATE-----
MIIBYjCCAQigAwIBAgIBATAKBggqhkjOPQQDAjAXMRUwEwYDVQQDEwxMRUFQIFJv
b3QgQ0EwHhcNMjExMTAyMTkwNTM3WhcNMjYxMTAyMTkxMDM3WjAXMRUwEwYDVQQD
EwxMRUFQIFJvb3QgQ0EwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAAQxOXBGu+gf
pjHzVteGTWL6XnFxtEnKMFpKaJkA/VOHmESzoLsZRQxt88GssxaqC01J17idQiqv
zgNpedmtvFtyo0UwQzAOBgNVHQ8BAf8EBAMCAqQwEgYDVR0TAQH/BAgwBgEB/wIB
ATAdBgNVHQ4EFgQUZdoUlJrCIUNFrpffAq+LQjnwEz4wCgYIKoZIzj0EAwIDSAAw
RQIgfr3w4tnRG+NdI3LsGPlsRktGK20xHTzsB3orB0yC6cICIQCB+/9y8nmSStfN
VUMUyk2hNd7/kC8nL222TTD7VZUtsg==
-----END CERTIFICATE-----`)

var riseupVPNCA = []byte(`-----BEGIN CERTIFICATE-----
MIIBYjCCAQigAwIBAgIBATAKBggqhkjOPQQDAjAXMRUwEwYDVQQDEwxMRUFQIFJv
b3QgQ0EwHhcNMjExMTAyMTkwNTM3WhcNMjYxMTAyMTkxMDM3WjAXMRUwEwYDVQQD
EwxMRUFQIFJvb3QgQ0EwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAAQxOXBGu+gf
pjHzVteGTWL6XnFxtEnKMFpKaJkA/VOHmESzoLsZRQxt88GssxaqC01J17idQiqv
zgNpedmtvFtyo0UwQzAOBgNVHQ8BAf8EBAMCAqQwEgYDVR0TAQH/BAgwBgEB/wIB
ATAdBgNVHQ4EFgQUZdoUlJrCIUNFrpffAq+LQjnwEz4wCgYIKoZIzj0EAwIDSAAw
RQIgfr3w4tnRG+NdI3LsGPlsRktGK20xHTzsB3orB0yC6cICIQCB+/9y8nmSStfN
VUMUyk2hNd7/kC8nL222TTD7VZUtsg==
-----END CERTIFICATE-----
`)
