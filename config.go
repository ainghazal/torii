package main

import (
	"log"

	"github.com/spf13/viper"
)

const (
	envPrefix = "TORII"
	appName   = "torii"
)

func loadConfig() {
	viper.SetEnvPrefix(envPrefix)
	viper.AutomaticEnv()
	viper.SetConfigName(appName)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	// not capturing error because config file is optional
	_ = viper.ReadInConfig()
}

func skipTLS() bool {
	return viper.Get("insecure") == "yes"
}

func serverName() string {
	sn := viper.Get("server_name")
	if !skipTLS() && sn == nil {
		log.Fatal("ERROR: missing server_name in config (tls: true)")
	}
	return sn.(string)
}
