package wireguard

import (
	"fmt"
	"log"
	"os/exec"
)

type ProcessManagerInterface interface {
	ReloadConfig() error
}

type ProcessManagerStub struct{}

func (pm *ProcessManagerStub) ReloadConfig() error {
	log.Printf("[STUB] Reloading config")
	return nil
}

type ProcessManager struct {
	InterfaceName string
}

func (pm *ProcessManager) ReloadConfig() error {
	script := fmt.Sprintf("wg syncconf %s <(wg-quick strip %s)", pm.InterfaceName, pm.InterfaceName)
	cmd := exec.Command("/bin/bash", "-c", script)
	_, err := cmd.Output()
	if err != nil {
		return err
	}
	return nil
}
