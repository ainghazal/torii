package vpn

import (
	"encoding/base64"
	"encoding/pem"
	"errors"
)

var (
	errNoKey        = errors.New("cannot decode key")
	errNoCert       = errors.New("cannot decode cert")
	typePrivateKey  = "RSA PRIVATE KEY"
	typeCertificate = "CERTIFICATE"
)

// toBase64 encodes a pem block as a base64: prefixed block.
func toBase64(b []byte) string {
	return "base64:" + base64.URLEncoding.EncodeToString(b)
}

func splitCombinedPEM(combined []byte) (key, cert []byte, err error) {
	key = []byte{}
	cert = []byte{}
	err = nil

	assignBlocks := func(block *pem.Block) {
		switch block.Type {
		case typePrivateKey:
			key = pem.EncodeToMemory(block)
		case typeCertificate:
			cert = pem.EncodeToMemory(block)
		}
	}

	block, rest := pem.Decode(combined)

	if block == nil || (block.Type != typeCertificate && block.Type != typePrivateKey) {
		err = errNoKey
		return key, cert, err
	}

	assignBlocks(block)

	block, rest = pem.Decode(rest)
	if block == nil || (block.Type != typeCertificate && block.Type != typePrivateKey) {
		err = errNoCert
		return key, cert, err
	}

	assignBlocks(block)

	return key, cert, err
}
