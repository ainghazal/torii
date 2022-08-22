package main

import (
	"log"
	"net/http"

	"github.com/spf13/viper"
	"golang.org/x/crypto/acme/autocert"
)

// autoTLSServer returns an http server configured with automatic LE certificates.
func autoTLSServer(r http.Handler) (*http.Server, autocert.Manager) {
	sn := serverName()
	log.Printf("Configuring TLS cert for %s\n", sn)
	email := ""
	if cfgEmail := viper.Get("email"); cfgEmail != nil {
		email = cfgEmail.(string)
	}

	certManager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(serverName()),
		Cache:      autocert.DirCache("certs"),
		Email:      email,
	}

	server := &http.Server{
		Addr:      ":https",
		TLSConfig: certManager.TLSConfig(),
		Handler:   r,
	}

	return server, certManager
}
