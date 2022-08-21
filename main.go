package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	bolt "go.etcd.io/bbolt"

	"github.com/ainghazal/torii/share"
	"github.com/ainghazal/torii/vpn"
)

var (
	listeningPort = ":8080"
)

const (
	authorName = "Ain Ghazal <ain@openobservatory.org>"

	paramProvider    = "provider"
	paramCountryCode = "cc"

	errNotFoundStr = "not found"
	errTryAgainStr = "try again later"
	errNoConfig    = "cannot build config"

	msgHomeStr = "nothing to see here"
)

type httpHandler func(http.ResponseWriter, *http.Request)

func renderConfigForProvider(provider vpn.Provider, selector providerSelectorFn) (*config, error) {
	endpoints := selector(provider)
	if len(endpoints) == 0 {
		return nil, errors.New(errNoConfig)
	}
	auth := provider.Auth()

	netTests := []netTest{}

	for _, endpoint := range endpoints {
		test := netTest{
			TestName: endpoint.Proto, // one of: openvpn, wg
			Inputs: []string{
				fmt.Sprintf(
					"vpn://%s.%s/?addr=%s:%s&transport=%s",
					provider.Name(),
					endpoint.Proto,
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
		netTests = append(netTests, test)
	}
	return &config{
		Name:        fmt.Sprintf("openvpn-%s", provider.Name()),
		Description: fmt.Sprintf("measure vpn connection to random %s gateways", provider.Name()),
		Author:      authorName,
		NetTests:    netTests,
	}, nil
}

func randomEndpointDescriptor(w http.ResponseWriter, r *http.Request) {
	providerName := getParam(paramProvider, r)
	if !vpn.IsKnownProvider(providerName) {
		http.Error(w, errNotFoundStr, http.StatusNotFound)
		return
	}
	p := vpn.Providers[providerName]
	cfg, err := renderConfigForProvider(p, randomEndpointPicker())
	if err != nil {
		http.Error(w, errorString(err), http.StatusGatewayTimeout)
		return
	}
	json.NewEncoder(w).Encode(cfg)
}

func byCountryEndpointDescriptor(w http.ResponseWriter, r *http.Request) {
	providerName := getParam(paramProvider, r)
	if !vpn.IsKnownProvider(providerName) {
		http.Error(w, errNotFoundStr, http.StatusNotFound)
		return
	}

	cc := getParam(paramCountryCode, r)

	p := vpn.Providers[providerName]
	cfg, err := renderConfigForProvider(p, byCountryEndpointPicker(cc, 1))
	if err != nil {
		http.Error(w, errorString(err), http.StatusGatewayTimeout)
		return
	}
	json.NewEncoder(w).Encode(cfg)
}

func DescriptorByUUIDHandler(db *bolt.DB) httpHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		uuid := getParam("uuid", r)
		exp := share.GetExperimentByUUID(db, uuid)[0]
		p := vpn.Providers[exp.Provider]
		cc := exp.CountryCode
		max := exp.Max

		cfg, err := renderConfigForProvider(p, byCountryEndpointPicker(cc, strToIntOrOne(max)))
		if err != nil {
			http.Error(w, errorString(err), http.StatusGatewayTimeout)
			return
		}
		json.NewEncoder(w).Encode(cfg)
	}
}

func strToIntOrOne(s string) int {
	maxInt := 1
	if s != "" {
		mi, err := strconv.Atoi(s)
		if err == nil {
			maxInt = mi
		}
	}
	return maxInt
}

func errorString(err error) string {
	if os.Getenv("DEBUG") == "1" {
		return err.Error()
	}
	return errTryAgainStr
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(msgHomeStr))
}

func getParam(param string, r *http.Request) string {
	return mux.Vars(r)[param]
}

func main() {
	rand.Seed(time.Now().UnixNano())
	db, err := share.InitDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	log.Println("🌿 Initializing all providers...")
	err = vpn.InitAllProviders()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("🚀 Starting web server at", listeningPort)

	r := mux.NewRouter().StrictSlash(false)
	r.HandleFunc("/", homeHandler)
	api := r.PathPrefix("/api").Subrouter()
	shr := r.PathPrefix("/share").Subrouter()
	vpn := r.PathPrefix("/vpn").Subrouter()

	// user pages
	shr.HandleFunc("/new", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./assets/new.html")
	})
	shr.HandleFunc("/list", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./assets/list.html")
	})

	// api calls
	api.HandleFunc("/experiment/add", share.AddExperimentHandler(db))
	api.HandleFunc("/experiment/list", share.ListExperimentHandler(db))
	api.HandleFunc("/experiment/{uuid}", share.RenderJSONExperimentByUUID(db))

	// json handlers
	vpn.HandleFunc("/random/{provider}.json", randomEndpointDescriptor)
	vpn.HandleFunc("/{cc}/{provider}.json", byCountryEndpointDescriptor)
	shr.HandleFunc("/{uuid}", DescriptorByUUIDHandler(db))

	log.Fatal(http.ListenAndServe(listeningPort, r))
}
