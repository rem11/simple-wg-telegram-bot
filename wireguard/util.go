package wireguard

import (
	"bytes"
	"fmt"
	"math/big"
	"net"
	"sort"
)

func ipToInt(ipAddr net.IP) *big.Int {
	ipInt := big.NewInt(0)
	ipInt.SetBytes(ipAddr.To16())
	return ipInt
}

func intToIp(ipInt *big.Int) net.IP {
	srcBytes := ipInt.Bytes()
	dstLen := net.IPv6len
	if len(srcBytes) <= net.IPv4len {
		dstLen = net.IPv4len
	}
	dstBytes := make([]byte, dstLen)
	copy(dstBytes[dstLen-len(srcBytes):], srcBytes)
	var result net.IP = dstBytes
	return result
}

func isBroadcast(addr net.IP, network net.IPNet) bool {
	netAddr := network.IP
	if len(network.Mask) == net.IPv4len {
		addr = addr.To4()
		netAddr = netAddr.To4()
	}
	for i := 0; i < len(network.Mask); i++ {
		if addr[i] != netAddr[i]+^network.Mask[i] {
			return false
		}
	}
	return true
}

func validate(addr net.IP, network net.IPNet) error {
	if !network.Contains(addr) {
		return fmt.Errorf("address %s doesn't belong to network %s", addr, network)
	}
	if network.IP.Equal(addr) {
		return fmt.Errorf("address %s is a network address for network %s", addr, network)
	}
	if isBroadcast(addr, network) {
		return fmt.Errorf("address %s is a broadcast address for network %s", addr, network)
	}
	return nil
}

func getNextIPAddress(addrList []net.IP, network net.IPNet) (net.IP, error) {
	// Validate all addresses
	for _, addr := range addrList {
		err := validate(addr, network)
		if err != nil {
			return nil, fmt.Errorf("invlid address list: %w", err)
		}
	}

	// Sort addresses in ascending order
	sort.Slice(addrList, func(i, j int) bool {
		return bytes.Compare(addrList[i], addrList[j]) < 0
	})

	// Add network address to address list
	completeAddrList := make([]net.IP, len(addrList)+1)
	completeAddrList[0] = network.IP
	copy(completeAddrList[1:], addrList)

	// Iterate over resulting list and try finding a gaps in it
	var intPrev big.Int
	for i, addr := range completeAddrList {
		addrInt := ipToInt(addr)
		if i >= 1 {
			diff := big.NewInt(0)
			diff.Sub(addrInt, &intPrev)
			if diff.Cmp(big.NewInt(1)) == 1 {
				intNext := big.NewInt(0)
				intNext.Add(&intPrev, big.NewInt(1))
				return intToIp(intNext), nil
			}
		}
		if i == len(completeAddrList)-1 {
			intNext := big.NewInt(0)
			intNext.Add(addrInt, big.NewInt(1))
			result := intToIp(intNext)
			err := validate(result, network)
			if err != nil {
				return nil, fmt.Errorf("can't get next address: %w", err)
			}
			return result, nil
		}
		intPrev = *addrInt
	}
	return nil, fmt.Errorf("this error shouldn't occur")
}
