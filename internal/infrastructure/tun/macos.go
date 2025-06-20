//go:build darwin

package tun

import (
	"os/exec"
	"strconv"

	"github.com/songgao/water"
)

type TunConfig struct {
	InterfaceName string
	IP            string
	MTU           int
	Network       string
}

type TunManager struct {
	Config TunConfig
}

func NewTunManager(config TunConfig) *TunManager {
	return &TunManager{Config: config}
}

func (t *TunManager) Create() (*water.Interface, error) {
	config := water.Config{
		DeviceType: water.TUN,
	}
	config.Name = t.Config.InterfaceName

	tunInterface, err := water.New(config)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("ifconfig", t.Config.InterfaceName, t.Config.IP, t.Config.IP, "up")
	_, err = cmd.Output()
	if err != nil {
		return nil, err
	}

	cmd = exec.Command("ifconfig", t.Config.InterfaceName, "mtu", strconv.Itoa(t.Config.MTU))
	_, err = cmd.Output()
	if err != nil {
		return nil, err
	}

	cmd = exec.Command("route", "add", "-net", t.Config.Network, "-interface", t.Config.InterfaceName)
	_, err = cmd.Output()
	if err != nil {
		return nil, err
	}

	return tunInterface, nil
}
