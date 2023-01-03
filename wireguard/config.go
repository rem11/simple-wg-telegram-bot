package wireguard

type Config struct {
	Interface
	Peer []Peer
}

type Interface struct {
	Address    string
	PrivateKey string
	ListenPort string
}

type Peer struct {
	AllowedIPs string
	PublicKey  string
	Name       string
}
