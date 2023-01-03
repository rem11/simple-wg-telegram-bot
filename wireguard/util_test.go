package wireguard

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TODO: rewrite like here https://github.com/go-ini/ini/blob/b2f570e5b5b844226bbefe6fb521d891f529a951/struct_test.go#L90

func TestIsBroadcast(t *testing.T) {
	t.Run("address is not a broadcast one", func(t *testing.T) {
		addr, network, _ := net.ParseCIDR("192.168.1.1/24")
		require.False(t, isBroadcast(addr, *network))
	})

	t.Run("address is a broadcast one", func(t *testing.T) {
		_, network, _ := net.ParseCIDR("192.168.1.1/24")
		addr := net.ParseIP("192.168.1.255")
		require.True(t, isBroadcast(addr, *network))
	})
}

func TestValidate(t *testing.T) {
	t.Run("valid address", func(t *testing.T) {
		addr, network, _ := net.ParseCIDR("192.168.1.1/24")
		require.NoError(t, validate(addr, *network))
	})

	t.Run("invalid: broadcast address", func(t *testing.T) {
		_, network, _ := net.ParseCIDR("192.168.1.1/24")
		addr := net.ParseIP("192.168.1.255")
		require.Error(t, validate(addr, *network))
	})

	t.Run("invalid: network address", func(t *testing.T) {
		_, network, _ := net.ParseCIDR("192.168.1.1/24")
		addr := net.ParseIP("192.168.1.0")
		assert.Error(t, validate(addr, *network))
	})

	t.Run("invalid: different network", func(t *testing.T) {
		_, network, _ := net.ParseCIDR("192.168.1.1/24")
		addr := net.ParseIP("192.168.2.1")
		assert.Error(t, validate(addr, *network))
	})

}

func TestNextIP(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		_, network, _ := net.ParseCIDR("192.168.1.0/24")
		addrList := []net.IP{}
		nextAddr, _ := getNextIPAddress(addrList, *network)
		require.Equal(t, nextAddr, net.ParseIP("192.168.1.1"))
	})

	t.Run("single IP in list", func(t *testing.T) {
		_, network, _ := net.ParseCIDR("192.168.1.0/24")
		addrList := []net.IP{
			net.ParseIP("192.168.1.1"),
		}
		nextAddr, _ := getNextIPAddress(addrList, *network)
		require.Equal(t, nextAddr, net.ParseIP("192.168.1.2"))
	})

	t.Run("single IP in list with free space before", func(t *testing.T) {
		_, network, _ := net.ParseCIDR("192.168.1.0/24")
		addrList := []net.IP{
			net.ParseIP("192.168.1.2"),
		}
		nextAddr, _ := getNextIPAddress(addrList, *network)
		require.Equal(t, nextAddr, net.ParseIP("192.168.1.1"))
	})

	t.Run("several IP in list with space inbetween", func(t *testing.T) {
		_, network, _ := net.ParseCIDR("192.168.1.0/24")
		addrList := []net.IP{
			net.ParseIP("192.168.1.1"),
			net.ParseIP("192.168.1.5"),
		}
		nextAddr, _ := getNextIPAddress(addrList, *network)
		require.Equal(t, nextAddr, net.ParseIP("192.168.1.2"))
	})

	t.Run("overflow with /31 network", func(t *testing.T) {
		_, network, _ := net.ParseCIDR("192.168.1.0/31")
		addrList := []net.IP{}
		_, err := getNextIPAddress(addrList, *network)
		require.Error(t, err)
	})

	t.Run("overflow with /30 network", func(t *testing.T) {
		_, network, _ := net.ParseCIDR("192.168.1.0/30")
		addrList := []net.IP{
			net.ParseIP("192.168.1.1"),
			net.ParseIP("192.168.1.2"),
		}
		_, err := getNextIPAddress(addrList, *network)
		require.Error(t, err)
	})

	t.Run("no overflow with /30 network", func(t *testing.T) {
		_, network, _ := net.ParseCIDR("192.168.1.0/30")
		addrList := []net.IP{
			net.ParseIP("192.168.1.1"),
		}
		nextAddr, _ := getNextIPAddress(addrList, *network)
		require.Equal(t, nextAddr, net.ParseIP("192.168.1.2"))
	})

}
