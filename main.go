package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
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
	TestName string   `json:"test_name"`
	Inputs   []string `json:"inputs"`
	Options  options  `json:"options"`
}

type options struct {
	Cipher   string
	Auth     string
	SafeCa   string
	SafeCert string
	SafeKey  string
}

type remote struct {
	IP        string
	Port      string
	Transport string
}

var providerRemotes map[string][]remote

func initProviderRemotes() {
	providerRemotes = make(map[string][]remote)

	// TODO fetch this from the provider API
	riseupRemotes := []remote{
		remote{"204.13.164.252", "1194", "udp"},
		remote{"204.13.164.252", "1194", "tcp"},
	}

	providerRemotes["riseup"] = riseupRemotes
}

func pickRandomRemote(provider string) remote {
	all := providerRemotes[provider]
	pick := rand.Intn(len(all))
	return all[pick]
}

func singleConfig(provider string) *config {

	p := vpn.Providers[provider]
	auth := p.Auth()
	r := pickRandomRemote(provider)

	test := netTest{
		TestName: "openvpn",
		Inputs: []string{
			fmt.Sprintf(
				"vpn://%s.openvpn/?addr=%s:%s&transport=%s",
				provider,
				r.IP,
				r.Port,
				r.Transport,
			)},
		Options: options{
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
	}
}

func riseupDescriptor(w http.ResponseWriter, r *http.Request) {
	cfg := singleConfig("riseup")
	json.NewEncoder(w).Encode(cfg)
}

func main() {
	rand.Seed(time.Now().UnixNano())
	initProviderRemotes()
	router := mux.NewRouter().StrictSlash(false)
	router.HandleFunc("/vpn/riseup.json", riseupDescriptor)
	log.Fatal(http.ListenAndServe(":8080", router))
}
