package wireguard

type ClientInterface struct {
	Address    string
	DNS        string
	PrivateKey string
}

type ClientPeer struct {
	PublicKey  string
	AllowedIPs string
	Endpoint   string
}

type ClientConfig struct {
	Interface ClientInterface
	Peer      ClientPeer
}
