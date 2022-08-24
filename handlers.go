package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/ainghazal/torii/share"
	"github.com/ainghazal/torii/vpn"
	"github.com/gorilla/mux"
	bolt "go.etcd.io/bbolt"
)

type httpHandler func(http.ResponseWriter, *http.Request)

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

// newCustomProviderFromExperiment returns a "custom" provider from a given
// experiment spec.
// This is a little bit hacky for the time being.
// the goal is to be able to provide a deterministic (set of) remotes, to be
// able to easily test and replicate anomalies etc. Only one remote now, coming
// from the form.
// we generate a provider "on the fly" and construct a single remote that is
// assoiated with it. The usage of some of the fields (like cc) is a little
// handwavy for now, but it serves my purpose well.
func newCustomProviderFromExperiment(exp *share.Experiment) vpn.Provider {
	// this assumes we're given a remote in the experiment definition.
	p := vpn.NewCustomProvider(exp.Provider)
	p.CustomName = exp.Provider + "-" + exp.Name

	if exp.Provider != "unknown" {
		// This is tricky, I need a registry mechanism to include
		// auth for my own endpoints. all of them should identify as
		// "unknown" for the experiment, but we need a handle
		// internally.
		refProvider := vpn.Providers[exp.Provider]
		p.AuthFromProvider(refProvider)
	}
	remoteParts := strings.Split(exp.EndpointRemote, ":")

	// TODO be defensive, validate
	ip := remoteParts[0]
	port := remoteParts[1]

	customEndpoint := &vpn.Endpoint{
		Label:       exp.Name,
		IP:          ip,
		Port:        port,
		Proto:       "openvpn", //only one for now
		Transport:   "tcp",
		Obfuscation: "none",
		CountryCode: exp.CountryCode, // this could be a wrong one, need to check against the canonical list
	}
	p.AddEndpoint(customEndpoint)
	return p
}

func DescriptorByUUIDHandler(db *bolt.DB) httpHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		uuid := getParam("uuid", r)
		exp := share.GetExperimentByUUID(db, uuid)[0]

		var cfg *config
		var err error
		var p vpn.Provider

		if exp.EndpointRemote != "" {
			p = newCustomProviderFromExperiment(exp)
			cfg, err = renderConfigForProvider(p, randomEndpointPicker())
		} else {
			p := vpn.Providers[exp.Provider]
			cc := exp.CountryCode
			max := exp.Max
			cfg, err = renderConfigForProvider(p, byCountryEndpointPicker(cc, strToIntOrOne(max)))
		}
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
