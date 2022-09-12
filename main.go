package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	health "github.com/ainghazal/health-check"
	"github.com/ainghazal/torii/share"
	"github.com/ainghazal/torii/vpn"
)

var (
	listeningPort                   = ":8080"
	providersWithEnabledHealthCheck = []string{"riseup"}
	healthServiceMap                = make(map[string]*health.HealthService)
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

func isEnabledProvider(name string) bool {
	return hasItem(providersWithEnabledHealthCheck, name)
}

func hasItem(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func main() {
	initRand()
	loadConfig()

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
	for name, provider := range vpn.Providers {
		if isEnabledProvider(name) {
			hs := &health.HealthService{
				Name: name,
				Checker: &health.VPNChecker{
					Provider: provider,
				},
			}
			hs.Start()
			healthServiceMap[name] = hs
		}
	}

	r := mux.NewRouter().StrictSlash(false)
	r.HandleFunc("/", homeHandler)
	api := r.PathPrefix("/api").Subrouter()
	shr := r.PathPrefix("/share").Subrouter()
	vpn := r.PathPrefix("/vpn").Subrouter()
	st := r.PathPrefix("/status").Subrouter()

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

	// status handlers
	st.HandleFunc("/riseup/status/json", health.HealthQueryHandlerJSON(healthServiceMap, "riseup")).Queries("addr", "{addr}").Queries("tr", "{tr}")
	st.HandleFunc("/riseup/summary", health.HealthSummaryHandlerText(healthServiceMap, "riseup"))

	if skipTLS() {
		log.Println("🚀 Starting web server at", listeningPort)
		log.Fatal(http.ListenAndServe(listeningPort, r))
	} else {
		tlsServer, certManager := autoTLSServer(http.Handler(r))
		go http.ListenAndServe(":http", certManager.HTTPHandler(nil))
		log.Fatal(tlsServer.ListenAndServeTLS("", ""))
	}
}
