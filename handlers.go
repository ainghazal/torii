package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"

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
