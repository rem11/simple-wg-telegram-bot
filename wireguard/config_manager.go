package wireguard

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"gopkg.in/ini.v1"
)

type ConfigManager struct {
	ConfigFilePath string
	Hostname       string
	DNS            string
	InterfaceName  string
	ProcessManager ProcessManagerInterface
}

func calculateNextIP(config *Config) (net.IP, error) {
	ifaceAddr, network, err := net.ParseCIDR(config.Interface.Address)
	if err != nil {
		return nil, fmt.Errorf("error parsing interface address: %w", err)
	}

	addrList := make([]net.IP, len(config.Peer)+1)
	addrList[0] = ifaceAddr
	for i, peer := range config.Peer {
		addr, _, err := net.ParseCIDR(peer.AllowedIPs)
		if err != nil {
			return nil, fmt.Errorf("error parsing peer AllowedIPs: %w", err)
		}
		addrList[i+1] = addr
	}

	nextIP, err := getNextIPAddress(addrList, *network)
	if err != nil {
		return nil, fmt.Errorf("error getting next ip address: %w", err)
	}

	return nextIP, nil
}

func (c *ConfigManager) loadConfig() (*ini.File, *Config, error) {
	cfgFile, err := ini.LoadSources(ini.LoadOptions{AllowNonUniqueSections: true}, c.ConfigFilePath)
	if err != nil {
		return nil, nil, fmt.Errorf("error loading config: %w", err)
	}

	config := &Config{}

	err = cfgFile.MapTo(config)
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing config: %w", err)
	}

	sections, err := cfgFile.SectionsByName("Peer")
	if err == nil {
		config.Peer = make([]Peer, len(sections))
		for i, section := range sections {
			err = section.MapTo(&config.Peer[i])
			if err != nil {
				return nil, nil, fmt.Errorf("error parsing peer: %w", err)
			}
			config.Peer[i].Name = strings.Trim(section.Comment, "# ")
		}
	}

	return cfgFile, config, nil
}

func (c *ConfigManager) AddPeer(publicKey string, name string) error {
	cfgFile, config, err := c.loadConfig()
	if err != nil {
		return err
	}

	for _, peer := range config.Peer {
		if peer.PublicKey == publicKey {
			return fmt.Errorf("peer with public key %s already exists: %s", publicKey, peer.Name)
		}
	}

	// Backup original config, so we could restore it in case something fails after we save it.
	cfgBackup := *cfgFile

	sec, err := cfgFile.NewSection("Peer")
	if err != nil {
		return fmt.Errorf("error creating section: %w", err)
	}

	_, err = sec.NewKey("PublicKey", publicKey)
	if err != nil {
		return fmt.Errorf("error adding PublicKey: %w", err)
	}

	sec.NewKey("PublicKey", publicKey)
	sec.Comment = "# " + name

	nextIP, err := calculateNextIP(config)
	if err != nil {
		return fmt.Errorf("error calculating next IP address for peer: %w", err)
	}

	sec.NewKey("AllowedIPs", nextIP.String()+"/32")

	err = cfgFile.SaveTo(c.ConfigFilePath)
	if err != nil {
		return fmt.Errorf("error saving configuration: %w", err)
	}

	err = c.ProcessManager.ReloadConfig()
	if err != nil {
		cfgBackup.SaveTo(c.ConfigFilePath)
		return fmt.Errorf("error reloading configration: %w", err)
	}

	return nil
}

func getPeerIndex(config *Config, publicKey string) (int, error) {
	for i, peer := range config.Peer {
		if peer.PublicKey == publicKey {
			return i, nil
		}
	}
	return -1, errors.New("can't find peer with specified public key")
}

func (c *ConfigManager) RemovePeer(publicKey string) error {
	cfgFile, config, err := c.loadConfig()
	if err != nil {
		return err
	}

	// Backup original config, so we could restore it in case something fails after we save it.
	cfgBackup := *cfgFile

	index, err := getPeerIndex(config, publicKey)
	if err != nil {
		return err
	}

	err = cfgFile.DeleteSectionWithIndex("Peer", index)
	if err != nil {
		return fmt.Errorf("error removing peer section: %w", err)
	}

	err = cfgFile.SaveTo(c.ConfigFilePath)
	if err != nil {
		return fmt.Errorf("error saving configuration: %w", err)
	}

	err = c.ProcessManager.ReloadConfig()
	if err != nil {
		cfgBackup.SaveTo(c.ConfigFilePath)
		return fmt.Errorf("error reloading configration: %w", err)
	}

	return nil
}

func (c *ConfigManager) ListPeers() ([]Peer, error) {
	_, config, err := c.loadConfig()
	if err != nil {
		return nil, err
	}

	return config.Peer, nil
}

func (c *ConfigManager) getClientConfigStruct(publicKey string) (*ClientConfig, error) {
	_, config, err := c.loadConfig()
	if err != nil {
		return nil, err
	}

	index, err := getPeerIndex(config, publicKey)
	if err != nil {
		return nil, err
	}

	addr, _, err := net.ParseCIDR(config.Peer[index].AllowedIPs)
	if err != nil {
		return nil, fmt.Errorf("error peer AllowdIPs: %w", err)
	}

	_, network, err := net.ParseCIDR(config.Interface.Address)
	if err != nil {
		return nil, fmt.Errorf("error parsing interface address: %w", err)
	}

	privateKey, err := wgtypes.ParseKey(config.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("error parsing private key: %w", err)
	}

	maskSize, _ := network.Mask.Size()

	return &ClientConfig{
		Interface: ClientInterface{
			PrivateKey: "<put your private key here>",
			Address:    addr.String() + "/" + strconv.Itoa(maskSize),
			DNS:        c.DNS,
		},
		Peer: ClientPeer{
			Endpoint:   c.Hostname + ":" + config.Interface.ListenPort,
			AllowedIPs: "0.0.0.0/0, ::/0",
			PublicKey:  privateKey.PublicKey().String(),
		},
	}, nil
}

func (c *ConfigManager) GetClientConfig(publicKey string) (*ClientConfig, string, error) {
	clientConfig, err := c.getClientConfigStruct(publicKey)
	if err != nil {
		return nil, "", err
	}

	cfgFile := ini.Empty()
	err = cfgFile.ReflectFrom(clientConfig)
	if err != nil {
		return nil, "", fmt.Errorf("error creating ini file from struct: %w", err)
	}

	buffer := bytes.NewBufferString("")
	_, err = cfgFile.WriteTo(buffer)
	if err != nil {
		return nil, "", fmt.Errorf("error writing ini contents to file: %w", err)
	}
	return clientConfig, buffer.String(), nil
}
