package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

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

	if skipTLS() {
		log.Println("🚀 Starting web server at", listeningPort)
		log.Fatal(http.ListenAndServe(listeningPort, r))
	} else {
		tlsServer, certManager := autoTLSServer(http.Handler(r))
		go http.ListenAndServe(":http", certManager.HTTPHandler(nil))
		log.Fatal(tlsServer.ListenAndServeTLS("", ""))
	}
}
