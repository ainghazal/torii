package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"

	"github.com/ainghazal/torii/vpn"
)

var (
	authorName = "Ain Ghazal <ain@openobservatory.org>"
)

type config struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Author      string    `json:"author"`
	NetTests    []netTest `json:"nettests"`
}

type netTest struct {
	TestName string      `json:"test_name"`
	Inputs   []string    `json:"inputs"`
	Options  vpn.Options `json:"options"`
}

func pickRandomEndpoint(provider string) *vpn.Endpoint {
	p := vpn.Providers[provider]
	all := p.Endpoints()
	if len(all) == 0 {
		return nil
	}
	pick := rand.Intn(len(all))
	log.Printf("Picked endpoint %d/%d\n", pick+1, len(all))
	return all[pick]
}

var errNoConfig = errors.New("cannot build config")

func singleVPNConfig(provider string) (*config, error) {
	p := vpn.Providers[provider]
	auth := p.Auth()
	endpoint := pickRandomEndpoint(provider)
	if endpoint == nil {
		return nil, errNoConfig
	}

	test := netTest{
		TestName: "openvpn",
		Inputs: []string{
			fmt.Sprintf(
				"vpn://%s.openvpn/?addr=%s:%s&transport=%s",
				provider,
				endpoint.IP,
				endpoint.Port,
				endpoint.Transport,
			)},
		Options: vpn.Options{
			Cipher:   "AES-256-GCM",
			Auth:     "SHA512",
			SafeCa:   auth.Ca,
			SafeCert: auth.Cert,
			SafeKey:  auth.Key,
		},
	}
	return &config{
		Name:        fmt.Sprintf("openvpn-%s", provider),
		Description: fmt.Sprintf("measure vpn connection to random %s gateways", provider),
		Author:      authorName,
		NetTests:    []netTest{test},
	}, nil
}

func errorString(err error) string {
	if os.Getenv("DEBUG") == "1" {
		return err.Error()
	}
	return "try again later"
}

func riseupDescriptor(w http.ResponseWriter, r *http.Request) {
	cfg, err := singleVPNConfig("riseup")
	if err != nil {
		http.Error(w, errorString(err), http.StatusGatewayTimeout)
		return
	}
	json.NewEncoder(w).Encode(cfg)
}

var listeningPort = ":8080"

func main() {
	rand.Seed(time.Now().UnixNano())

	log.Println("🌿 Initializing all providers...")
	err := vpn.InitAllProviders()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("🚀 Starting web server at", listeningPort)

	router := mux.NewRouter().StrictSlash(false)
	router.HandleFunc("/vpn/riseup.json", riseupDescriptor)
	log.Fatal(http.ListenAndServe(listeningPort, router))
}
