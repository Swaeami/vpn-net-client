//go:build windows

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

	tunInterface, err := water.New(config)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("netsh", "interface", "ip", "set", "address",
		"name="+t.Config.InterfaceName, "static", t.Config.IP, "255.255.255.0")
	_, err = cmd.Output()
	if err != nil {
		return nil, err
	}

	cmd = exec.Command("netsh", "interface", "ipv4", "set", "subinterface",
		t.Config.InterfaceName, "mtu="+strconv.Itoa(t.Config.MTU), "store=persistent")
	_, err = cmd.Output()
	if err != nil {
		return nil, err
	}

	cmd = exec.Command("route", "add", t.Config.Network, "mask", "255.255.255.0", t.Config.IP)
	_, err = cmd.Output()
	if err != nil {
		return nil, err
	}

	return tunInterface, nil
}
