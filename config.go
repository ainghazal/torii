package main

import (
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
