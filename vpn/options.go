package vpn

// Options are the options for a vpn nettest
// TODO split for wireguard and openvpn.
type Options struct {
	Cipher   string
	Auth     string
	SafeCa   string
	SafeCert string
	SafeKey  string
}
