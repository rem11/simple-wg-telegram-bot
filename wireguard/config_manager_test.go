package wireguard

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const testConfig = `[Interface]
Address    = 192.168.3.1/24
ListenPort = 11111
PrivateKey = sLsJoF6gLXYWfRcpRkA7ugzvkYX15Lpvif5oBeZeaHA=
`

const testClientConfig = `[Interface]
Address    = 192.168.3.2/24
DNS        = 8.8.8.8
PrivateKey = <put your private key here>

[Peer]
PublicKey  = V5CyX8fiyVYu9R4qZelHaJl915y6jsUDwlbT/abgOVY=
AllowedIPs = 0.0.0.0/0, ::/0
Endpoint   = example.com:11111
`

func prepareTestConfig() (string, error) {
	file, err := os.CreateTemp(".", "test-config")
	if err != nil {
		return "", fmt.Errorf("can't create temporary file: %w", err)
	}
	defer file.Close()
	_, err = file.WriteString(testConfig)
	if err != nil {
		return "", fmt.Errorf("can't write temporary file: %w", err)
	}
	return file.Name(), nil
}

func TestConfigManager(t *testing.T) {
	configFile, err := prepareTestConfig()
	require.NoError(t, err)
	defer os.Remove(configFile)

	configManager := ConfigManager{
		ConfigFilePath: configFile,
		ProcessManager: &ProcessManagerStub{},
		DNS:            "8.8.8.8",
		Hostname:       "example.com",
	}

	err = configManager.AddPeer("yyy", "Test Peer")
	require.NoError(t, err)

	peers, _ := configManager.ListPeers()
	require.NotEmpty(t, peers)
	require.Equal(t, peers[0].Name, "Test Peer")
	require.Equal(t, peers[0].PublicKey, "yyy")
	require.Equal(t, peers[0].AllowedIPs, "192.168.3.2/32")

	clientConfig, configStr, err := configManager.GetClientConfig("yyy")
	require.NoError(t, err)
	require.Equal(t, clientConfig.Interface.Address, "192.168.3.2/24")
	require.Equal(t, clientConfig.Interface.DNS, "8.8.8.8")
	require.Equal(t, clientConfig.Peer.Endpoint, "example.com:11111")
	require.Equal(t, clientConfig.Peer.PublicKey, "V5CyX8fiyVYu9R4qZelHaJl915y6jsUDwlbT/abgOVY=")
	require.Equal(t, configStr, testClientConfig)

	err = configManager.RemovePeer("yyy")
	require.NoError(t, err)

	peers, _ = configManager.ListPeers()
	require.Empty(t, peers)
}
