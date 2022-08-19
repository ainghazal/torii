package main

import "github.com/ainghazal/torii/vpn"

type config struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Author      string    `json:"author"`
	NetTests    []netTest `json:"nettests"`
}

type netTest struct {
	TestName string   `json:"test_name"`
	Inputs   []string `json:"inputs"`
	// TODO these options can be generalized via an interface
	Options vpn.Options `json:"options"`
}
